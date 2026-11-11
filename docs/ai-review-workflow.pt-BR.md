# Workflow de revisão por IA (opt-in)

[English](ai-review-workflow.md) · **Português (Brasil)**

[`ai-review.yml`](../.github/workflows/ai-review.yml) roda a revisão multiagente da skill
[`pr-review-process`](../.agents/skills/pr-review-process/SKILL.md) com `anthropics/claude-code-action`.
É um conselho e não bloqueia nada. Ele não roda o motor determinístico; para um gate que bloqueia, rode o
[`action.yml`](../action.yml) da raiz com `fail-on-block: true` como check obrigatório.

## Como ligar

1. Adicione o secret de repositório `ANTHROPIC_API_KEY`. Sem ele, o job imprime uma mensagem de skip e passa.
2. Mantenha o `REVIEW.md` atualizado. Ele é o contrato do revisor (profundidade, limite de nits, caminhos ignorados, evidências).
   O revisor lê o arquivo do branch base, então uma mudança vale depois do merge.
3. Opcionalmente, rode à mão: Actions, ai-review, Run workflow, com o número de um PR. Uma execução manual
   publica só o resumo: a action fixada instala a ferramenta de comentário inline apenas em eventos de pull request.

Ele nunca roda em PRs de fork nem em drafts. As permissões são `contents: read`, mais `pull-requests: write` e
`issues: write` para comentários e labels.

## Rodadas

| Rodada | O que o revisor lê |
|---|---|
| 1 | O diff inteiro contra o repositório todo e a intenção do PR |
| 2, 3 | Só os commits desde o último SHA revisado; um achado já publicado nunca é repetido |
| 4+ | Sem revisão. O workflow pede que o autor reconsidere a abordagem |

O estado fica no PR: o label `ai-review:round-N` e um comentário-marcador oculto com o último SHA
revisado. Apague os labels para zerar a contagem. Os relatórios terminam com `Round N of 3`, que o
[`scripts/review-noise.sh`](../scripts/review-noise.sh) lê.

## Segurança

O diff em revisão é escrito pelo autor do PR, então qualquer coisa nele pode tentar conduzir o revisor.
O workflow trata o PR inteiro como dado e mantém segredos e instruções fora do alcance dele.

- **Sem shell.** Steps confiáveis rodam antes do Claude e gravam as entradas dele em arquivos: o diff
  (`git diff --no-ext-diff --no-textconv`) e o título, o corpo e as mensagens de commit do PR. O Claude recebe
  só `Read`, `Grep`, `Glob`, subagentes e a ferramenta de comentário inline da action; `Bash`, `Edit`, `Write`
  e as ferramentas web são negadas. Só números, SHAs validados e caminhos do runner entram no prompt.
- **Leituras ficam no checkout.** O `--restricted` confina as ferramentas de arquivo ao checkout e à pasta de
  entradas, e regras de negação também bloqueiam `/proc`, a pasta temporária do runner e todo `.git/` (a action
  grava o token dela no `.git/config`). O `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` liga a limpeza de segredos, em
  melhor esforço, que a CLI faz nos subprocessos.
- **Saída confiável.** O Claude devolve o relatório como saída estruturada. Um step posterior lê esse relatório
  via `env:`, mascara strings com formato de token e publica. O Claude não tem ferramenta que escreva
  comentários no PR além dos inline.
- **Configuração e instruções do base.** Antes de o Claude começar, `.claude/`, `.mcp.json`, `.claude.json`,
  `CLAUDE.md`, `CLAUDE.local.md`, `.gitmodules`, `.ripgreprc` e `.husky` voltam ao commit base (são apagados se
  o base não os tiver), e a CLI roda com `--setting-sources user --strict-mcp-config`, então um PR não consegue
  adicionar hooks nem servidores MCP. O revisor segue `SKILL.md`, `REVIEW.md`, `AGENTS.md` e `reviewers/` de
  uma worktree do commit base, então um PR não consegue reescrever a própria revisão.
- **Estado das rodadas.** O último SHA revisado vem só de um comentário cujo corpo inteiro bate com o marcador
  deste workflow, e só é usado se ainda for ancestral do head; senão, a rodada revisa o PR inteiro.
- **O que sobra.** Texto injetado ainda pode enviesar achados ou publicar um comentário inline enganoso, e um
  segredo codificado passaria pelo mascaramento se alguma leitura escapasse. A revisão é um conselho; uma pessoa decide.
- A action está fixada em um SHA completo (v1.0.271); as entradas, saídas e flags dela foram conferidas nessa
  versão. Confira de novo antes de atualizar.

## Custo e ruído

Rode `scripts/review-noise.sh owner/repo --since YYYY-MM-DD` todo mês
([process/03](../process/03-noise-metrics.pt-BR.md)). As classes dele são heurísticas, então amostre.

## Evals noturnos de incidentes

[`review-evals.yml`](../.github/workflows/review-evals.yml) roda `go run ./cmd/bench` toda noite e
falha se verdadeiros positivos, omissões, falsos positivos, o silêncio em casos limpos ou qualquer nota de classe regredirem
em relação ao `bench/results.json`. Adicione um caso ao corpus para cada incidente e atualize a linha de base com
`make bench` no mesmo PR ([process/04](../process/04-incident-evals.pt-BR.md)). Ele pontua o motor determinístico,
não o revisor LLM.
