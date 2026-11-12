# Classificador de nível de risco

**Português (Brasil)** · [English](review-tier.en.md)

O `cmd/tier` lê um diff e imprime o nível de risco (tier) e quantas pessoas precisam aprovar. É Go puro
(biblioteca padrão), determinístico, e nunca chama um modelo. A mudança define o tier, não quem a escreveu:
[process/01-risk-tiers.md](../process/01-risk-tiers.md).

## Uso

```bash
go run ./cmd/tier --from origin/main --to HEAD            # text
go run ./cmd/tier --from origin/main --to HEAD --format json --author someone
git diff main... | go run ./cmd/tier -                    # stdin
go run ./cmd/tier change.diff                             # file
go run ./cmd/tier --incomplete "motivo"                    # change set ilegível: high
```

Exit 0 classificado, 2 erro de uso, de leitura, de git ou de configuração (uma configuração inválida nunca cai em low). JSON:

```json
{"schemaVersion":1,"tier":"medium","humanApprovals":1,"alwaysHuman":false,"approvalMode":"enforce",
 "reasons":[{"path":"internal/x.go","rule":"medium: internal/**","tier":"medium"}]}
```

## Regras

| Regra | Efeito |
|---|---|
| O caminho mais alto vence | O tier do PR é o do caminho alterado mais arriscado (renomeações e remoções contam, os dois lados) |
| Tier para pessoas | low 0, medium 1, high 2 |
| Nenhuma regra casa | `default` da configuração (defina como `medium`) |
| Globs `alwaysHuman` e `.github/workflows/**` | `alwaysHuman: true`, no mínimo 1 pessoa, qualquer que seja o tier |
| `--author` fora de `trustedAuthors` | No mínimo medium. Ignorado quando a lista está vazia ou nenhum autor é informado |
| Diff vazio | low, 0 pessoas |
| `--incomplete MOTIVO` | high, 2 pessoas, o motivo fica registrado; nenhum diff é lido (falha fechada) |

Globs: `*` e `?` ficam dentro de um segmento de caminho, `**` é um segmento inteiro que casa com zero ou mais segmentos,
e um padrão sem `/` casa com o nome-base em qualquer profundidade (`*.md`). `docs/**` é tudo que está abaixo de
`docs/`. Uma regra `*.md` (low) perde para `security/skills/**` (medium) em `security/skills/x/SKILL.md`, porque o tier mais
alto vence. A configuração rejeita campos desconhecidos e globs inválidos.

Configuração: [`.github/review-tiers.json`](../.github/review-tiers.json); quem adota copia
`templates/review-tiers.json`.

## Aplicação (ligue isto)

[`review-tier.yml`](../.github/workflows/review-tier.yml) aplica ao PR o label `risk:*` (e `needs-human`),
publica o tier e roda um job chamado **`human-approvals`**. Esse job é um **sinal consultivo**: mostra
quantas aprovações o tier pede e quantas o **commit head atual** tem. Ele não impõe nada (veja abaixo).
Só contam revisores com permissão admin, maintain ou write (checada por revisor na API de permissão de
colaboradores; se a chamada falhar, o revisor não conta e o log avisa). Logins de bot, contas `[bot]` e o autor do PR nunca contam; um "changes requested"
posterior da mesma pessoa cancela a aprovação dela. Ele roda de novo a cada revisão. Sempre executa e
falha fechado: se o `classify` foi cancelado, falhou ou não produziu tier, ele falha, exceto no caso
explícito de bootstrap abaixo.

**A imposição é a branch protection nativa ou um ruleset**, com "Require approvals", "Require review from
Code Owners" e "Dismiss stale approvals". Não dependa do `human-approvals` para isso: no evento
`pull_request_review` o GitHub executa o arquivo de workflow do merge ref do PR (inclusive em forks),
então qualquer PR pode adicionar um workflow próprio com um job chamado `human-approvals`, e os checks
obrigatórios não distinguem checks com o mesmo nome.

Só uma pessoa mantém o repo? O GitHub não deixa o autor aprovar o próprio PR, então um PR medium ou high
nunca passaria. Coloque `"approvals": "advisory"` na config e o job avisa em vez de falhar; o tier, as
labels e o comentário continuam iguais. O padrão (`"enforce"`, ou sem o campo) falha. Este repo usa
`advisory` porque tem um único mantenedor.

O workflow usa `pull_request_target` e lê o classificador e a configuração da **branch base**, então
um PR não consegue baixar o próprio tier. Ele nunca executa código do PR e nunca busca nem faz checkout da head do PR: os caminhos alterados vêm
da API REST do GitHub (lista de arquivos do PR) como dados. Se a API não entrega a lista completa (erro,
limite de 3000 arquivos, contagem divergente, caracteres de controle num caminho) o PR é classificado
**high** com o motivo registrado, nunca low. Quando a branch base ainda não tem `cmd/tier` (o PR que o
introduz), o `classify` define `bootstrap=true`, emite um aviso e pula a classificação, e o
`human-approvals` passa com um aviso. Nos eventos `pull_request_target`, o workflow e o classificador vêm
da branch base, então um PR não os altera nesse evento; o evento `pull_request_review` é diferente
(acima). A saída do classificador é conferida contra uma lista de valores permitidos (tier, humans,
always, mode) antes do uso; qualquer outra coisa falha o job.

Configurações do repositório (Settings, Rules ou Branches):

1. Exija um pull request com o número de aprovações que o tier pede (1 ou 2) e exija os status checks
   `go` e `review`. O `human-approvals` pode constar também, mas é um sinal, não o portão.
2. **Descarte aprovações obsoletas quando chegarem commits novos.**
3. **Exija revisão dos Code Owners** ([`CODEOWNERS`](../.github/CODEOWNERS) lista os caminhos always-human e high).
4. Exija uma merge queue, para que as checagens rodem de novo contra a main com a mudança aplicada. Adicione o
   gatilho `merge_group` a qualquer workflow que precise rodar ali (hoje o `ci.yml` só tem `push` e `pull_request`).
5. Não dê acesso de escrita nem direito de aprovação ao bot do Actions nem a uma conta de agente. Um agente
   que aprova com uma conta de usuário normal com acesso de escrita não é detectado; mantenha essas contas
   fora do repositório ou adicione os logins delas ao filtro jq do workflow.

Limites: o classificador só enxerga caminhos; a segunda opinião (`reviewers/risk-second-opinion.md`) cobre
o que os caminhos não cobrem. "Quem deu o prompt de um PR escrito por agente conta como um revisor" é uma regra humana
registrada no template de PR, não algo que o workflow consiga detectar.
