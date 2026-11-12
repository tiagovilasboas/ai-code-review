# Pranchas

**Português (Brasil)** · [English](README.en.md)

Oito pranchas em estilo de gravura e um cartão social para o README e o site. Cada uma desenha uma parte do processo de revisão em sete camadas que o time do Claude Code na Anthropic descreveu em [*How the Claude Code team does code review*](https://claude.dev/blog/solving-code-review-at-anthropic/), e cada uma credita esse post na cartela de fonte e no `<desc>` do SVG. Este repositório é uma implementação independente, sem vínculo ou endosso da Anthropic.

As pranchas são geradas. Não edite um SVG à mão: mude os textos ou o gerador e gere de novo.

## As pranchas

Cada prancha tem um arquivo claro e um escuro, em inglês aqui e em português em [`pt-BR/`](pt-BR/). O cartão social só existe no tema escuro.

| Nº | Português | English | Arquivos | Tamanho |
|---|---|---|---|---|
| I | As Sete Camadas | The Seven Layers | `hero-{light,dark}.svg` | 1280 × 600 |
| II | As Camadas em Corte | The Layers in Section | `seven-layers-{light,dark}.svg` | 1280 × 800 |
| III | O Fluxo da Revisão | The Review Pipeline | `review-pipeline-{light,dark}.svg` | 1280 × 600 |
| IV | Níveis de Risco | Risk Tiers | `risk-tiers-{light,dark}.svg` | 1280 × 780 |
| V | Convergência | Convergence | `convergence-{light,dark}.svg` | 1280 × 840 |
| VI | Qualquer Harness | Any Harness | `harness-map-{light,dark}.svg` | 1280 × 800 |
| VII | O Ciclo de Aprendizado | The Learning Loop | `learning-loop-{light,dark}.svg` | 1280 × 820 |
| VIII | A Borda da Cloudflare | The Cloudflare Edge | `cloudflare-edge-{light,dark}.svg` | 1280 × 800 |
| cartão | Cartão social | Social card | `og-dark.svg` | 1280 × 640 |

O build também copia as pranchas II, III, VI e VIII para [`site/art/`](../../site/art/), e o cartão em inglês para `site/og.svg` e `site/og.png`.

## Vinhetas sem borda

Seis gravuras sem moldura, sem legenda de prancha e sem fundo, uma para cada seção do README. O papel fica transparente e, onde uma forma precisa cobrir o que está atrás, usa a cor da página do GitHub (`#ffffff` no claro, `#0d1117` no escuro). No gerador elas usam os temas `BARE` (`bare: true` em `plates/index.js`).

| Seção do README | Arquivos | Tamanho | Fatos e origem |
|---|---|---|---|
| Por que um revisor determinístico | `v-deterministic-{light,dark}.svg` | 1280 × 400 | Mesmo diff, mesmo veredito, sem LLM nem rede ([`README.md`](../../README.md)) |
| Início rápido | `v-quickstart-{light,dark}.svg` | 1280 × 400 | O comando e a saída real do CLI para `security/examples/sample-pr.diff`, com exit 1 |
| O que ela pega | `v-catches-{light,dark}.svg` | 1280 × 400 | As famílias de regras de [`security/runbook.md`](../../security/runbook.md) |
| Instale em qualquer harness | `v-install-{light,dark}.svg` | 1280 × 424 | Os seis harnesses de [`harness/`](../../harness/README.md) |
| Como funciona | `v-engine-{light,dark}.svg` | 1280 × 396 | Parse, regras, relatório; severidade, `arquivo:linha` e motivo |
| Guardrails | `v-guardrails-{light,dark}.svg` | 1280 × 446 | Os três guardrails de [`reviewers/guardrails/`](../../reviewers/guardrails/) |

## O que cada prancha afirma e de onde vem

As pranchas só afirmam fatos que o repositório já documenta. Quando um fato muda, mude primeiro o arquivo de origem e depois os textos.

| Prancha | Fatos | Origem no repositório |
|---|---|---|
| I | As sete camadas; um defeito barrado na III e outro na V | [`process/00-seven-layers.md`](../../process/00-seven-layers.md) |
| II | O que implementa cada camada aqui; menos revisão humana nunca pode significar mais reverts ou incidentes | [`process/00-seven-layers.md`](../../process/00-seven-layers.md) |
| III | As oito estações; 1, 3 ou 5 hunters conforme a profundidade; publica com 2 de 3 (deep 3 de 3); limite de 3 nits; "Round N of 3"; aprovação de agente nunca conta | [`.agents/skills/pr-review-process/SKILL.md`](../../.agents/skills/pr-review-process/SKILL.md) |
| IV | Low 0, medium 1, high 2 aprovações; o caminho mais alto vence; a segunda opinião só aumenta; arquivos de workflow de CI exigem uma pessoa; aprovação de bot não conta; aprovações descartadas em commits novos | [`process/01-risk-tiers.md`](../../process/01-risk-tiers.md), [`docs/review-tier.md`](../review-tier.md) |
| V | Mais de 50 rodadas e cerca de 59% já presentes; 65% para 78%, 8,9% para 1,5%, 2,08 para 1,32; limite de 3 rodadas | [`process/02-convergence.md`](../../process/02-convergence.md) (números relatados pelos autores do post) |
| VI | Seis harnesses; processo completo no Claude Code, sequencial nos outros; a mesma votação, o mesmo limite de rodadas e o mesmo BLOCK em todos; o comando de instalação | [`harness/README.md`](../../harness/README.md), `scripts/install-review-process.sh` |
| VII | Incidente, qual camada, regra, caso do bench, pontuado contra todos os casos; `make bench` sobre `bench/corpus` | [`process/04-incident-evals.md`](../../process/04-incident-evals.md), [`process/00-seven-layers.md`](../../process/00-seven-layers.md) § 7 |
| VIII | O que o CLI pega na configuração da Cloudflare, com severidade e ação; o que fica com o revisor; os IDs OWASP | [`security/skills/cloudflare-config.md`](../../security/skills/cloudflare-config.md), [`internal/cli/testdata/golden/cloudflare-config.sample.txt`](../../internal/cli/testdata/golden/cloudflare-config.sample.txt) |
| cartão | O recibo: `Decision: DO NOT MERGE`, `wrangler.toml:9`, `CRITICAL`, `BLOCK` | [`internal/cli/testdata/golden/cloudflare-config.sample.txt`](../../internal/cli/testdata/golden/cloudflare-config.sample.txt) |

## Usar uma prancha

Use `<picture>` para a prancha acompanhar o tema de cores de quem lê. A partir da raiz do repositório:

```html
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/art/pt-BR/hero-dark.svg">
  <img src="docs/art/pt-BR/hero-light.svg" width="1280" alt="Prancha de título gravada: AI Code Review, sete camadas de revisão para qualquer harness de agente. Sete fatias de queijo suíço em fila; um defeito é barrado na camada III, segurança que bloqueia, e outro na camada V, aprovação humana proporcional ao risco.">
</picture>
```

No `README.en.md` em inglês, aponte para `docs/art/` e use o texto alternativo em inglês. A partir de um arquivo em `process/`, o caminho é `../docs/art/pt-BR/`.

## Texto alternativo

| Prancha | Texto alternativo |
|---|---|
| I | Prancha de título gravada: AI Code Review, sete camadas de revisão para qualquer harness de agente. Sete fatias de queijo suíço em fila; um defeito é barrado na camada III, segurança que bloqueia, e outro na camada V, aprovação humana proporcional ao risco. |
| II | Prancha gravada das sete camadas empilhadas como lajes, cada uma com seu papel neste repositório. Um defeito cai pelos furos das camadas I e II e é barrado na camada III, segurança que bloqueia. |
| III | Fluxo gravado de oito estações: preparar, classificar o tier, o motor de segurança (que pode parar a execução com BLOCK), bug hunters em paralelo, a votação dos verificadores, um filtro, o relatório e uma pessoa que decide. |
| IV | Prancha gravada de três pedestais para os tiers de risco: low não exige aprovação humana quando a revisão por IA e a segurança estão verdes, medium exige uma, high exige duas. A mudança define o tier, não quem a escreveu. |
| V | Prancha gravada sobre convergência: uma espiral se fecha em três rodadas até "reconsidere a abordagem", ao lado de barras com os resultados relatados pelos autores do post (de 65% para 78% aprovados após uma rodada). |
| VI | Planetário gravado: o processo de revisão, definido uma vez nos reviewers e na skill, no centro; claude-code, kiro, codex, cursor, copilot e gemini no anel em volta. |
| VII | Roda gravada de cinco estações em volta de make bench: um incidente, qual camada deveria ter pego, a resposta vira uma regra, um caso do bench, e toda mudança pontuada contra todos os casos. |
| VIII | Portão gravado com a grade baixada, barrando o diff. À esquerda, o que o CLI pega, com severidade e ação; à direita, o que fica com o revisor. |
| cartão | Cartão gravado: AI Code Review, sete camadas de revisão para qualquer harness de agente. Um recibo de relatório diz Decision: DO NOT MERGE e Action: BLOCK; um defeito passa pelas fatias I e II e é barrado na III. |

Cada SVG também tem `role="img"`, um `<title>` e um `<desc>` longo que descreve todos os rótulos da prancha e a fonte.

## Estilo

- **Gravura.** Cada tom é um campo de linhas paralelas desenhado como `<pattern>` de SVG: cinco níveis do claro ao escuro, cruzados nos níveis 4 e 5, no ângulo que acompanha a forma. Sombras são pontilhado, contornos são traços que engrossam, e o papel é o único preenchimento chapado.
- **Paleta.** Claro: papel `#f4efe4`, tinta `#1d1a16`, destaque `#b3412c`. Escuro: papel `#14130f`, tinta `#e9e2d0`, destaque `#e0735a`. O destaque marca só o que bloqueia.
- **Tipografia.** EB Garamond, convertida em contornos para as pranchas não dependerem de fonte web. Legendas usam versaletes simulados; notas e linhas de figura são itálicas.
- **Composição.** Moldura com rosetas nos cantos, a legenda da prancha no topo, uma linha de figura e uma cartela de fonte embaixo. Pranchas de uma só camada acrescentam `Ref. §` e o título da seção do post, em inglês nas duas línguas.
- **Sem marcas.** Só rótulos de texto: nada de logos, logotipos ou cores de marca. Anthropic e Cloudflare aparecem só como texto gravado simples.

## Gerar de novo

O gerador é uma pequena ferramenta de build só para estas imagens; o motor de revisão em si é Go. Ele precisa de Node 18 ou mais recente, das dependências fixadas em [`generator/package.json`](generator/package.json) e do `rsvg-convert` para renderizar `site/og.png`.

```bash
npm --prefix docs/art/generator ci
npm --prefix docs/art/generator run build
npm --prefix docs/art/generator run build -- --preview /tmp/plates   # também um PNG de cada SVG
```

O build é determinístico: os mesmos textos geram os mesmos bytes. Cada SVG fica abaixo de 200 KB. Os textos ficam em [`generator/src/strings/en.json`](generator/src/strings/en.json) e [`generator/src/strings/pt-BR.json`](generator/src/strings/pt-BR.json); o código de desenho fica em `generator/src/pen`, `figures` e `plates`.

## Fontes

- **Processo de revisão.** Addy Osmani, Morgan Westlee Lunt e Qing Wang, [*How the Claude Code team does code review*](https://claude.dev/blog/solving-code-review-at-anthropic/), Anthropic, 15 de outubro de 2026. https://claude.dev/blog/solving-code-review-at-anthropic/
- **Cloudflare (prancha VIII).** developers.cloudflare.com: [Workers secrets](https://developers.cloudflare.com/workers/configuration/secrets/), [Access JWT validation](https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/) e [WAF](https://developers.cloudflare.com/waf/).
- **Tipografia.** EB Garamond, Copyright 2017 The EB Garamond Project Authors (https://github.com/octaviopardo/EBGaramond12), sob a SIL Open Font License 1.1. A licença está em [`generator/fonts/OFL-EB-Garamond.txt`](generator/fonts/OFL-EB-Garamond.txt).

Pranchas desenhadas em código por Tiago Vilas Boas. Texto e desenhos são deste repositório; erros são nossos, não dos autores do post.
