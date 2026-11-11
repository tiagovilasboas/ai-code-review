# Evals de incidentes

[English](04-incident-evals.md) · **Português (Brasil)**

Todo incidente que a revisão deveria ter pego vira um caso de regressão. O revisor roda contra o conjunto inteiro todo dia; qualquer mudança de prompt, modelo ou ferramenta é pontuada contra todos os casos e **só é mantida se a nota subir**.

## O padrão já existe aqui
O motor determinístico é construído assim. [`bench/corpus/<case-id>/`](../bench/README.md) contém:

| Arquivo | Conteúdo |
|---|---|
| `change.diff` | O unified diff que foi revisado |
| `expected.json` | `description`, `source` e `findings: [{path, line, class}]`, onde `line` é a linha no arquivo novo |

Um array `findings` vazio significa que o revisor deve ficar em silêncio (caso de PR limpo). `make bench` (`go run ./cmd/bench`, código em `internal/bench`) pontua o motor contra o corpus e grava `bench/results.json`. Os rótulos refletem o julgamento do revisor, nunca a saída atual do motor.

## Transforme um incidente em um caso
1. Ache o PR que introduziu o problema e o fix-forward ou o revert.
2. **Uma pessoa engenheira confirma** o local do bug (`path:line` no diff original). Sem confirmação, sem caso.
3. Salve o diff original como `change.diff`; escreva o `expected.json` com o local confirmado e uma `class`.
4. Adicione também o irmão limpo (o mesmo código com a checagem presente, `findings: []`), para que uma correção que aponta demais seja pega.
5. Rode o bench. Um caso novo falhando é justamente o objetivo: corrija a regra, o prompt ou a skill até ele passar sem quebrar os outros.
6. Registre na revisão do incidente qual camada deveria ter pego o problema.

## Para revisores LLM
Use o mesmo formato de corpus para os prompts de `reviewers/*`: rode cada papel sobre todo `change.diff` e compare os achados com o `expected.json` por `path:line`. Gate para uma mudança:

| Mudança | Aceitar somente se |
|---|---|
| Edição de prompt, modelo, ferramenta ou skill | O recall no conjunto inteiro não cai **e** o teste de PR limpo não piora |
| Caso novo adicionado | Ele falha primeiro, depois passa com a correção |

A pontuação nunca edita o `expected.json` para casar com a saída.

## Alternativa sem custo
Dispense o LLM: o motor determinístico sobre `bench/corpus` é grátis e roda em segundos. Para os papéis de LLM, rode o conjunto toda semana à mão em um harness que você já tem.

## Modos de falha
| Falha | Proteção |
|---|---|
| Rótulos não verificados | Uma pessoa engenheira confirma o local; `source` diz quem |
| Overfitting ao conjunto | Mantenha casos de `bench/blind/` sem ver enquanto ajusta |
| O conjunto apodrece | Execução diária; um caso que deixa de falhar pelo motivo certo é revisado |

Próximo: [05 papéis humanos](05-human-roles.pt-BR.md).
