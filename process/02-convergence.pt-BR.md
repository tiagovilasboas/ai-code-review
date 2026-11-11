# Convergência: pare a revisão que nunca termina

[English](02-convergence.md) · **Português (Brasil)**

**O problema** (números do post do time do Claude Code na Anthropic, [*How the Claude Code team does code review*](https://claude.dev/blog/solving-code-review-at-anthropic/)). Um PR teve mais de 50 rodadas de revisão durante a noite. Cerca de 59% do que foi apontado já estava lá na primeira passada: o revisor continuava achando coisas antigas sob uma luz nova, e quem escrevia (um agente) agia sobre cada uma. Este repositório é uma implementação independente desse processo, sem vínculo ou endosso da Anthropic.

## Controles
| Controle | Regra |
|---|---|
| Limite de rodadas | Depois de **3** rodadas o revisor para de listar e pede que o autor reconsidere a abordagem |
| Revisão só do diff | Depois da primeira passada, revise só o que mudou desde o último commit revisado |
| Classe de bug | Aponte a classe ("todos os handlers aqui pulam a checagem de dono"), não uma linha por vez |
| Sem repetição | Nunca republique um achado já publicado, resolvido ou respondido com um motivo |
| Varredura noturna | O código existente é revisado por uma execução agendada, não empurrado para dentro de PRs sem relação |

Implementado por [`reviewers/convergence.md`](../reviewers/convergence.md) e pela skill orquestradora ([`pr-review-process`](../.agents/skills/pr-review-process/SKILL.md)). Bugs preexistentes recebem o rótulo `Pre-existing`, nunca `Important`, e nunca bloqueiam o PR.

## Resultados reportados na origem
| Métrica | Antes | Depois |
|---|---|---|
| PRs aprovados após uma rodada | 65% | 78% |
| PRs que precisam de 5+ rodadas | 8,9% | 1,5% |
| Média de revisões por PR | 2,08 | 1,32 |

São números da equipe de origem, nos repositórios dela. Meça os seus ([03](03-noise-metrics.pt-BR.md)); não os assuma.

## Registro de rodadas
Todo relatório de revisão termina com `Round N of 3`. A rodada 3 termina com a nota de reconsiderar a abordagem, listando as classes ainda abertas e mais nada. Sem rodada 4 sem uma pessoa.

## Varredura noturna (sem custo)
Um job de CI agendado (ou cron em qualquer máquina) roda `go run ./cmd/review --git --from <last-sweep> --to HEAD` mais o bug hunter nos diretórios não tocados recentemente. Os achados vão para uma issue, não para dentro de um PR aberto.

## Modos de falha
| Falha | Proteção |
|---|---|
| O revisor derrapa para estilo | Achados de estilo e de mera limpeza são descartados (limite de nits em [`REVIEW.md`](../REVIEW.md)) |
| O agente autor entra em loop corrigindo nits | Limite de nits, limite de rodadas, regra de não repetir |
| O limite esconde um bug real | Um BLOCK de segurança fica fora do limite; permanece até ser corrigido ou resolvido com um motivo |

Próximo: [03 ruído e métricas](03-noise-metrics.pt-BR.md).
