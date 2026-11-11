# Workflow de revisão por IA (opt-in)

[English](ai-review-workflow.md) · **Português (Brasil)**

[`ai-review.yml`](../.github/workflows/ai-review.yml) roda a revisão multiagente da skill
[`pr-review-process`](../.agents/skills/pr-review-process/SKILL.md) com `anthropics/claude-code-action`.
É um conselho em cima do gate determinístico; a checagem de segurança que bloqueia continua sendo o job `review` do `ci.yml`.

## Como ligar

1. Adicione o secret de repositório `ANTHROPIC_API_KEY`. Sem ele, o job imprime uma mensagem de skip e passa.
2. Mantenha o `REVIEW.md` atualizado. Ele é o contrato do revisor (profundidade, limite de nits, caminhos ignorados, evidências).
3. Opcionalmente, rode à mão: Actions, ai-review, Run workflow, com o número de um PR.

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

- O título, o corpo e o diff do PR são dados não confiáveis. Só números e SHAs validados entram no prompt.
- A lista de ferramentas é somente leitura mais comentários inline; o agente não consegue fazer merge, aprovar nem editar.
- A action está fixada em um SHA completo (v1.0.271). Confira as entradas dela antes de atualizar.
- Custo e ruído são com você: rode `scripts/review-noise.sh owner/repo --since YYYY-MM-DD`
  todo mês ([process/03](../process/03-noise-metrics.pt-BR.md)). As classes dele são heurísticas, então amostre.

## Evals noturnos de incidentes

[`review-evals.yml`](../.github/workflows/review-evals.yml) roda `go run ./cmd/bench` toda noite e
falha se verdadeiros positivos, omissões, falsos positivos, o silêncio em casos limpos ou qualquer nota de classe regredirem
em relação ao `bench/results.json`. Adicione um caso ao corpus para cada incidente e atualize a linha de base com
`make bench` no mesmo PR ([process/04](../process/04-incident-evals.pt-BR.md)). Ele pontua o motor determinístico,
não o revisor LLM.
