# As sete camadas do code review

[English](00-seven-layers.md) · **Português (Brasil)**

Nenhuma verificação pega tudo. Cada camada tem furos; empilhe as camadas para que os furos não se alinhem (o modelo do "queijo suíço"). A pergunta para cada mudança não é "quem revisa isto?", e sim **"o que pode dar errado nesta mudança, e qual verificação a pega melhor?"**

Métrica de proteção para o processo inteiro: **menos revisão humana nunca pode significar mais reverts ou incidentes.** Acompanhe reverts e fix-forwards (correções aplicadas em um novo commit, sem reverter) por PR mesclado antes e depois de cada afrouxamento. Se subirem, aperte de novo.

| # | Camada | Implementada aqui por | Alternativa sem custo |
|---|---|---|---|
| 1 | Verificar antes do PR | [`reviewers/pre-pr-self-check.md`](../reviewers/pre-pr-self-check.md) | Rode lint e testes você mesmo, cole a saída, releia o próprio diff amanhã |
| 2 | Revisão por IA em todo PR | [`reviewers/bug-hunter.md`](../reviewers/bug-hunter.md), [`reviewers/verifier.md`](../reviewers/verifier.md), [`.agents/skills/pr-review-process/SKILL.md`](../.agents/skills/pr-review-process/SKILL.md), integração com harnesses em [`harness/`](../harness/README.pt-BR.md) | Rode a skill em qualquer harness que você já paga, ou revise à mão com o checklist |
| 3 | Segurança que bloqueia | [`reviewers/security.md`](../reviewers/security.md) mais o motor determinístico `cmd/review` ([`action.yml`](../action.yml), [`ai-review.yml`](../.github/workflows/ai-review.yml)) | O motor não precisa de LLM nem de rede |
| 4 | Checagens automáticas que bloqueiam | Seu CI; o motor é uma checagem obrigatória | Qualquer CI mais as perguntas do template de PR |
| 5 | Aprovação humana proporcional ao risco | `cmd/tier`, [`reviewers/risk-classifier.md`](../reviewers/risk-classifier.md), [`review-tier.yml`](../.github/workflows/review-tier.yml), CODEOWNERS | `go run ./cmd/tier` é determinístico; não precisa de modelo |
| 6 | Fila de merge e release em etapas | Sua plataforma (merge queue, flags) | Exija "branch atualizada" antes do merge |
| 7 | Detectar rápido, fazer rollback | Suas flags e o plantão | Um comando de revert documentado e um responsável |

As evidências para aprender com incidentes ficam em [`bench/`](../bench/README.md) (veja [04](04-incident-evals.pt-BR.md)).

## 1. Verificar antes de o PR existir
Quem escreve (pessoa ou agente) roda testes, build e lint e **mostra a saída**. Depois, um subagente com **contexto novo** revisa só o diff: o raciocínio de quem escreveu contamina a própria revisão. Passes opcionais antes do PR: caça a bugs, simplificação, verificação ponta a ponta.
- Aqui: `reviewers/pre-pr-self-check.md`.
- Alternativa: cole a saída dos testes no PR; um colega (ou uma sessão nova) lê o diff sem contexto prévio.

## 2. Revisão por IA em todo PR
Leia o diff contra o repositório inteiro **e a intenção declarada do PR**: ele resolve o problema declarado, cabe no sistema, quebra outra coisa? Em correção de bug, procure a causa raiz, não o sintoma. Vários caçadores (hunters) trabalham em paralelo; **verificadores** separados votam em cada achado (finding); só é publicado o achado em que verificadores suficientes concordam. Achados de estilo, de diagnóstico e de mera limpeza são descartados. Rótulos: `Important`, `Nit`, `Pre-existing`. Profundidade por repositório: `quick | standard | deep` (definida em [`REVIEW.md`](../REVIEW.md)).
- Aqui: papéis `bug-hunter` e `verifier`, orquestrados pela skill `pr-review-process`; limites de convergência em [02](02-convergence.pt-BR.md).
- Alternativa: um passe de hunter e um de verifier no mesmo harness; descarte o que o verifier não consegue reproduzir a partir do diff.

## 3. Revisão de segurança que pode bloquear
Um PR não pode ser mesclado enquanto houver um achado de segurança ou de bug em aberto. Corrija, ou resolva a conversa **com um motivo**; um commit posterior que corrige o problema limpa o achado.
- Aqui: o motor determinístico bloqueia com exit code 1 (HIGH/CRITICAL, `path:line`, somente IDs oficiais; o único HIGH que fica em REVIEW é a regra do header do Cloudflare Access, porque a checagem do JWT pode estar fora do diff). `reviewers/security.md` acrescenta julgamento de LLM, mas não pode anular um BLOCK.
- Alternativa: só o motor. É o gate de segurança sem custo e sem deriva.

## 4. Checagens automáticas que bloqueiam
Lint, tipos, testes em todas as plataformas, build. O template de PR pergunta **como foi testado e como desfazer**. Status checks obrigatórios, não sugestões.

## 5. Aprovação humana proporcional ao risco
A mudança define o nível (tier), não quem a escreveu. Veja [01](01-risk-tiers.pt-BR.md).

## 6. Fila de merge e release em etapas
A fila roda as checagens de novo contra a `main` com a mudança aplicada. Publique atrás de flags, primeiro para uso interno, depois em etapas (latest antes de stable).

## 7. Detecção rápida e rollback
Flags desligam sem deploy; alguém está de plantão. Toda revisão de incidente pergunta **qual camada deveria ter pego isso**, e a resposta vira uma regra, uma linha de skill ou um caso do bench ([04](04-incident-evals.pt-BR.md)).

Próximo: [01 níveis de risco](01-risk-tiers.pt-BR.md).
