# O que as pessoas ainda fazem

**Português (Brasil)** · [English](05-human-roles.en.md)

A IA revisa o diff; as pessoas são donas do resultado. O objetivo das outras camadas é gastar a atenção humana onde ela rende.

| Responsabilidade | Na prática |
|---|---|
| Decidir se vale construir | Antes de qualquer código: isso vale a pena? A revisão não salva uma ideia errada |
| Revisar código de risco | Cobrança, auth, auto-update, migrations, **workflows de CI** (sempre, por regra) |
| Revisar arquitetura que os outros vão copiar | Um padrão novo vira o modelo das próximas dez mudanças |
| Ler evidências, não resumos | Saída de testes, comandos executados, screenshots, o `path:line` do achado. Não o "todos os testes passam" do agente |
| Ser dono do resultado | Você aprovou, você é dono |

## Responsabilidade sem culpa
"Você mesclou, você é dono" significa que você corrige e escreve a revisão do incidente. Não significa que você leva a culpa. A revisão pergunta **qual camada deveria ter pego isso** e então muda o processo: uma regra de lint, uma linha de skill, uma linha no AGENTS.md/CLAUDE.md, um caso do bench ([04](04-incident-evals.md)).

## Para quem deu o prompt em um PR escrito por agente
Você conta como um revisor obrigatório ([01](01-risk-tiers.md)). Antes de pedir revisão a outras pessoas:
- Rode a mudança você mesmo; não encaminhe um diff que ninguém executou.
- Leia a saída da checagem pré-PR, não a conclusão dela.
- Saiba explicar a abordagem sem o agente.

## Para revisores
| Faça | Não faça |
|---|---|
| Abra a evidência que o relatório cita | Aprovar só pelo resumo do relatório |
| Resolva uma conversa com um motivo | Resolver só para destravar |
| Diga o que você não verificou | Sugerir cobertura total |
| Aprove o commit final | Confiar numa aprovação de um commit anterior |

## Escalonamento
BLOCK de segurança, tier `high` ou rodada 3 sem convergência: uma pessoa decide. Veja [`security/runbook.md`](../security/runbook.md) e [`reviewers/guardrails/write-approval.md`](../reviewers/guardrails/write-approval.md). O agente nunca faz merge.

Próximo: [06 checklist de adoção](06-adoption-checklist.md).
