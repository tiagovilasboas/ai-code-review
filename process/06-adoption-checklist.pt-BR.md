# Checklist de adoção

[English](06-adoption-checklist.md) · **Português (Brasil)**

## Antes do PR
- [ ] Testes, build e lint executados; saída mostrada (não resumida)
- [ ] Autoverificação do diff com contexto novo ([`pre-pr-self-check`](../reviewers/pre-pr-self-check.md))
- [ ] A descrição do PR traz a intenção, como foi testado e como desfazer

## Todo PR
- [ ] Tier calculado (`go run ./cmd/tier`) e publicado com os motivos
- [ ] Motor de segurança determinístico executado; BLOCK resolvido ou explicado
- [ ] Revisão por IA: hunters, depois verifiers; achados rotulados `Important | Nit | Pre-existing`
- [ ] O relatório começa com uma contagem; limite de 3 rodadas respeitado
- [ ] CI obrigatório verde em todas as plataformas

## Aprovação
- [ ] O número de pessoas corresponde ao tier (0 / 1 / 2); workflow de CI e caminhos `alwaysHuman` têm uma pessoa
- [ ] Quem aprova leu a evidência, não o resumo
- [ ] A aprovação está no commit final; aprovações obsoletas descartadas
- [ ] Aprovações de bot não contam; quem deu o prompt conta

## Depois do merge
- [ ] A fila de merge rodou as checagens de novo na `main`; flag ligada primeiro para uso interno, release em etapas
- [ ] Caminho de revert conhecido; responsável de plantão
- [ ] Houve incidente? Adicione um caso ao bench ([04](04-incident-evals.pt-BR.md)) e diga qual camada deveria ter pego
- [ ] Achado repetido? Corrija na origem: regra de lint, skill ou uma linha no AGENTS.md. Falso positivo repetido? Uma linha no `REVIEW.md`

## Rollout: comece em um repositório, ajuste por semanas, depois amplie
| Passo | O que fazer |
|---|---|
| 1. Instalar | `scripts/install-review-process.sh` em um repositório; copie [`templates/REVIEW.md`](../templates/REVIEW.md) e [`templates/review-tiers.json`](../templates/review-tiers.json) |
| 2. Observar | Rode os revisores só em modo de comentário; sem gate além do motor de segurança |
| 3. Ajustar | Toda semana: [`scripts/review-noise.sh`](../scripts/review-noise.sh), corrija no `REVIEW.md` os principais falsos positivos |
| 4. Travar | Ligue as checagens de tier e as aprovações obrigatórias quando o teste de PR limpo não mostrar relatórios prejudiciais |
| 5. Ampliar | Próximo repositório, com o `REVIEW.md` e os tiers dele. Nunca copie tiers às cegas |

## Choque de realidade
O processo de origem vem de uma equipe revisando o próprio código público e amplamente conhecido, algo próximo do melhor caso para um revisor de IA. Trate os números dela como um teto, não como uma previsão. Estudos apontam na mesma direção:

| Estudo | Resultado |
|---|---|
| CRBugBench (arXiv 2610.13668) | Agentes de revisão acharam o bug nos 5 primeiros comentários em 71% dos casos em repositórios open source, bem menos em repositórios fechados |
| TestSuiteBench (arXiv 2610.13065) | Suítes de teste escritas por agentes passam, mas pegam menos de 60% dos bugs plantados |
| Estudo de correções posteriores (arXiv 2609.26847) | PRs de agentes precisam de cerca de 1,5x mais correções posteriores; agentes corrigem cerca de 70% das próprias |

**Código aberto vs fechado:** seu código privado não está nos dados de treino de nenhum modelo, então espere uma revisão mais fraca do que a dos benchmarks públicos. Por isso o conjunto de evals de incidentes ([04](04-incident-evals.pt-BR.md)) é construído com **os seus** incidentes, e por isso o motor determinístico, que não depende da familiaridade do modelo com o código, continua sendo o gate de segurança bloqueante. Meça por conta própria; não aceite as afirmações na confiança.
