# Ruído e métricas

[English](03-noise-metrics.md) · **Português (Brasil)**

Ordem de prioridade: **primeiro pegar bugs reais, depois reduzir rodadas.** O ruído custa mais com autores que são agentes, porque um agente age sobre cada nit e o PR fica mudando sem parar.

## Classifique cada resposta do autor
| Classe | Significado | O que fazer |
|---|---|---|
| fixed | O autor mudou o código por causa do achado | Verdadeiro positivo; mantenha |
| said wrong | O autor mostrou que o achado está incorreto | Falso positivo; se repetir, adicione uma linha de "pare de apontar" ao `REVIEW.md` |
| intentional | O comportamento é deliberado | Documente a intenção no AGENTS.md ou em um comentário no código |
| didn't matter | Correto, mas irrelevante | Candidato a nit; suba o nível mínimo de severidade |

## Teste de PR limpo
Rode o revisor em PRs sabidamente limpos (já publicados, sem fix-forward nem revert depois). Conte os relatórios que levariam a uma **mudança prejudicial** se um agente os aplicasse. Meta: zero; qualquer ocorrência é uma regra a apertar.

## Cálculo
[`scripts/review-noise.sh`](../scripts/review-noise.sh) contabiliza as quatro classes, as rodadas por PR e os achados por PR a partir das conversas de revisão. Veja o cabeçalho do script para as entradas; ele é somente leitura e não precisa de LLM.

## Sinais-alvo
| Sinal | Direção saudável |
|---|---|
| Fatia dos achados classificados como `fixed` | Sobe |
| Fatia de `said wrong` + `didn't matter` | Desce |
| Aprovações após uma rodada | Sobe |
| PRs com 5+ rodadas | Desce |
| Média de revisões por PR | Desce, mas nunca à custa da linha seguinte |
| Reverts e fix-forwards por PR mesclado | **Estável ou em queda.** Se subir, aperte |

## Modos de falha
| Falha | Proteção |
|---|---|
| Otimizar rodadas e perder bugs | Meça primeiro os bugs pegos; o eval de incidentes ([04](04-incident-evals.pt-BR.md)) é o gate |
| Contar nits como vitórias | Só `fixed` conta como sinal |
| Métricas manipuladas fechando conversas | Resolver exige um motivo; amostre as resoluções todo mês |

Próximo: [04 evals de incidentes](04-incident-evals.pt-BR.md).
