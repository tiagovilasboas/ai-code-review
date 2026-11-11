# Níveis de risco e aprovação humana

[English](01-risk-tiers.md) · **Português (Brasil)**

**A mudança define o nível (tier), não quem a escreveu.** Um classificador determinístico atribui o tier a partir dos caminhos alterados; uma segunda opinião pergunta se esta mudança precisa de uma pessoa, qualquer que seja o tier.

| Tier | Mudanças típicas | Aprovações humanas |
|---|---|---|
| low | Docs, texto, código só interno, comportamento aditivo atrás de uma flag | 0, quando a revisão por IA e a segurança estiverem verdes |
| medium | Mudanças comuns de produto; áreas sensíveis como cache, auto-updater, system prompts | 1 |
| high | Código de maior risco (cobrança, auth, migrations) mais um pequeno conjunto de caminhos que sempre exigem duas | 2 |

## Regras
| Regra | Efeito |
|---|---|
| O caminho mais alto vence | Um único caminho high torna o PR inteiro high |
| Segunda opinião | [`reviewers/risk-second-opinion.md`](../reviewers/risk-second-opinion.md) só pode **aumentar** a necessidade de uma pessoa, nunca reduzir um tier |
| Globs `alwaysHuman` | Caminhos sensíveis marcados pelo dono sempre exigem uma pessoa, mesmo com tier low |
| Arquivos de workflow de CI | `.github/workflows/**` sempre exige uma pessoa |
| Autores confiáveis | O caminho leve (low, 0 humanos) vale só para logins em `trustedAuthors`. Autor não confiável significa no mínimo medium, 1 humano |
| Aprovação de bot nunca conta | Aprovações de agentes e bots não são aprovações humanas |
| Quem acionou o agente conta | A pessoa que deu o prompt de um PR escrito por agente é um revisor obrigatório, não um espectador |
| Vale para o commit | As aprovações são descartadas em commits novos; aprove de novo o head final |

## Aplicação neste kit
- **Classificar:** `go run ./cmd/tier --from origin/main --to HEAD --config .github/review-tiers.json` imprime
  `{"schemaVersion":1,"tier":"medium","humanApprovals":1,"alwaysHuman":false,"approvalMode":"enforce","reasons":[{"path":"...","rule":"...","tier":"..."}]}`. Exit 0 ok, 2 erro de uso ou de leitura. Todo tier tem `reasons` com caminhos, então um tier nunca fica sem explicação.
- **Reportar:** [`.github/workflows/review-tier.yml`](../.github/workflows/review-tier.yml) roda o `cmd/tier` e publica o tier e os motivos no PR.
- **Tornar obrigatório:** CODEOWNERS para os caminhos marcados pelo dono; branch protection com aprovações obrigatórias (1 ou 2) e **descartar aprovações obsoletas em novos commits**. A ferramenta de tier diz o número; a branch protection o impõe.
- **Configuração:** [`.github/review-tiers.json`](../.github/review-tiers.json) aqui; quem adota começa de [`templates/review-tiers.json`](../templates/review-tiers.json).

## Alternativa sem custo
O `cmd/tier` é Go puro, sem modelo. Sem rulesets de branch protection (repositórios privados em alguns planos), trate o comentário do tier como um item manual obrigatório do checklist.

## Modos de falha
| Falha | Proteção |
|---|---|
| Tier baixo demais | Segunda opinião, `alwaysHuman`, `medium` como padrão para caminhos sem regra |
| Abuso da confiança no autor | A confiança é uma lista por repositório, revisada como código |
| Aprovações carimbadas | Humanos leem as evidências ([05](05-human-roles.pt-BR.md)), não o resumo do agente |

Próximo: [02 convergência](02-convergence.pt-BR.md).
