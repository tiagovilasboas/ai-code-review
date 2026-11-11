# Classificador de nível de risco

[English](review-tier.md) · **Português (Brasil)**

O `cmd/tier` lê um diff e imprime o nível de risco (tier) e quantas pessoas precisam aprovar. É Go puro
(biblioteca padrão), determinístico, e nunca chama um modelo. A mudança define o tier, não quem a escreveu:
[process/01-risk-tiers.pt-BR.md](../process/01-risk-tiers.pt-BR.md).

## Uso

```bash
go run ./cmd/tier --from origin/main --to HEAD            # text
go run ./cmd/tier --from origin/main --to HEAD --format json --author someone
git diff main... | go run ./cmd/tier -                    # stdin
go run ./cmd/tier change.diff                             # file
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

Globs: `*` e `?` ficam dentro de um segmento de caminho, `**` é um segmento inteiro que casa com zero ou mais segmentos,
e um padrão sem `/` casa com o nome-base em qualquer profundidade (`*.md`). `docs/**` é tudo que está abaixo de
`docs/`. Uma regra `*.md` (low) perde para `security/skills/**` (medium) em `security/skills/x/SKILL.md`, porque o tier mais
alto vence. A configuração rejeita campos desconhecidos e globs inválidos.

Configuração: [`.github/review-tiers.json`](../.github/review-tiers.json); quem adota copia
`templates/review-tiers.json`.

## Aplicação (ligue isto)

[`review-tier.yml`](../.github/workflows/review-tier.yml) aplica ao PR o label `risk:*` (e `needs-human`),
publica o tier e roda um job chamado **`human-approvals`** que falha até que pessoas suficientes tenham aprovado o
**commit head atual**. Logins de bot, contas `[bot]` e o autor do PR nunca contam; um "changes requested"
posterior da mesma pessoa cancela a aprovação dela. Ele roda de novo a cada revisão.

Só uma pessoa mantém o repo? O GitHub não deixa o autor aprovar o próprio PR, então um PR medium ou high
nunca passaria. Coloque `"approvals": "advisory"` na config e o job avisa em vez de falhar; o tier, as
labels e o comentário continuam iguais. O padrão (`"enforce"`, ou sem o campo) falha. Este repo usa
`advisory` porque tem um único mantenedor.

O workflow usa `pull_request_target` e lê o classificador e a configuração da **branch base**, então
um PR não consegue baixar o próprio tier. Ele nunca executa código do PR. Mudanças no próprio workflow só valem
depois do merge.

Configurações do repositório (Settings, Rules ou Branches):

1. Exija um pull request e exija os status checks `human-approvals`, `go`, `review`.
2. **Descarte aprovações obsoletas quando chegarem commits novos.**
3. Exija revisão dos Code Owners ([`CODEOWNERS`](../.github/CODEOWNERS) lista os caminhos always-human e high).
4. Exija uma merge queue, para que as checagens rodem de novo contra a main com a mudança aplicada. Adicione o
   gatilho `merge_group` a qualquer workflow que precise rodar ali (hoje o `ci.yml` só tem `push` e `pull_request`).
5. Não dê direito de aprovação ao bot do Actions nem a uma conta de agente. Um agente que aprova com uma
   conta de usuário normal não é detectado; mantenha essas contas fora do repositório ou adicione os logins delas ao
   filtro jq do workflow.

Limites: o classificador só enxerga caminhos; a segunda opinião (`reviewers/risk-second-opinion.md`) cobre
o que os caminhos não cobrem. "Quem deu o prompt de um PR escrito por agente conta como um revisor" é uma regra humana
registrada no template de PR, não algo que o workflow consiga detectar.
