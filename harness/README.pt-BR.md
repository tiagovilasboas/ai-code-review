# Adaptadores de harness

[English](README.md) · **Português (Brasil)**

O processo de revisão é definido uma vez só, em `reviewers/*.md` (papéis) e `.agents/skills/pr-review-process/SKILL.md` (orquestrador). Cada pasta aqui é um adaptador fino: ela apenas aponta o harness para esses arquivos. Nenhuma lógica de papel é duplicada.

Instale com `scripts/install-review-process.sh --harness <name|all> --target <repo>` (ou `.ps1` no Windows). Veja o README de cada pasta.

## Matriz de capacidades

| Harness | Subagentes | Hunters em paralelo | Slash command | Hooks | Resultado |
|---|---|---|---|---|---|
| claude-code | sim (`.claude/agents`) | sim | `/review-pr`, `/pre-pr` | sim (exemplo opcional) | processo completo |
| kiro | não declarativo | não | gatilho manual por hook | sim (`agentStop`, manual) | sequencial, contexto limpo entre os papéis |
| codex | não declarativo | não (sessões separadas são possíveis) | prompt customizado (`.codex/prompts/`, copie para `~/.codex/prompts/`) | não | sequencial |
| cursor | não declarativo (background agents) | manual | não | não | sequencial |
| copilot | não declarativo | não | arquivo de prompt `/review-pr` | não | sequencial |
| gemini | não declarativo | não | `/review-pr` | não | sequencial |

## Como degrada

Quando um harness não consegue rodar subagentes novos e em paralelo, o orquestrador executa os papéis um após o outro e limpa o contexto entre eles (sessão nova, ou levando adiante só o diff, a intenção do PR e a lista de candidatos). O processo, a regra de votação (2 de 3; deep 3 de 3), o limite de 3 rodadas e o BLOCK de segurança continuam iguais. Só a independência fica mais fraca, e o relatório informa que a revisão rodou de forma sequencial.
