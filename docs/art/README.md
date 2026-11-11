# Plates

**English** · [Português (Brasil)](README.pt-BR.md)

Eight engraved plates and a social card for the README and the site. Each one draws a part of the seven-layer review process that the Claude Code team at Anthropic described in [*How the Claude Code team does code review*](https://claude.dev/blog/solving-code-review-at-anthropic/), and each one credits that post in its source cartouche and in its SVG `<desc>`. This repo is an independent implementation, not affiliated with or endorsed by Anthropic.

The plates are generated. Do not edit an SVG by hand: change the strings or the generator and rebuild.

## The plates

Every plate has a light and a dark file, in English here and in Portuguese under [`pt-BR/`](pt-BR/). The social card is dark only.

| No. | English | Português | Files | Size |
|---|---|---|---|---|
| I | The Seven Layers | As Sete Camadas | `hero-{light,dark}.svg` | 1280 × 600 |
| II | The Layers in Section | As Camadas em Corte | `seven-layers-{light,dark}.svg` | 1280 × 800 |
| III | The Review Pipeline | O Fluxo da Revisão | `review-pipeline-{light,dark}.svg` | 1280 × 600 |
| IV | Risk Tiers | Níveis de Risco | `risk-tiers-{light,dark}.svg` | 1280 × 780 |
| V | Convergence | Convergência | `convergence-{light,dark}.svg` | 1280 × 840 |
| VI | Any Harness | Qualquer Harness | `harness-map-{light,dark}.svg` | 1280 × 800 |
| VII | The Learning Loop | O Ciclo de Aprendizado | `learning-loop-{light,dark}.svg` | 1280 × 820 |
| VIII | The Cloudflare Edge | A Borda da Cloudflare | `cloudflare-edge-{light,dark}.svg` | 1280 × 800 |
| card | Social card | Cartão social | `og-dark.svg` | 1280 × 640 |

The build also copies plates II, III, VI and VIII to [`site/art/`](../../site/art/), and the English card to `site/og.svg` and `site/og.png`.

## What each plate states, and where it comes from

The plates only state facts that the repo already documents. When a fact changes, change it in the source file first, then in the strings.

| Plate | Facts | Source in this repo |
|---|---|---|
| I | The seven layers; a defect stopped at III and another at V | [`process/00-seven-layers.md`](../../process/00-seven-layers.md) |
| II | What implements each layer here; less human review must never mean more reverts or incidents | [`process/00-seven-layers.md`](../../process/00-seven-layers.md) |
| III | The eight stations; 1, 3 or 5 hunters by depth; post at 2 of 3 (deep 3 of 3); nit cap 3; "Round N of 3"; agent approvals never count | [`.agents/skills/pr-review-process/SKILL.md`](../../.agents/skills/pr-review-process/SKILL.md) |
| IV | Low 0, medium 1, high 2 approvals; highest path wins; the second opinion only raises; CI workflow files need a person; bot approvals do not count; approvals dismissed on new commits | [`process/01-risk-tiers.md`](../../process/01-risk-tiers.md), [`docs/review-tier.md`](../review-tier.md) |
| V | 50+ rounds and about 59% already there; 65% to 78%, 8.9% to 1.5%, 2.08 to 1.32; round cap 3 | [`process/02-convergence.md`](../../process/02-convergence.md) (numbers reported by the post's authors) |
| VI | Six harnesses; full process in Claude Code, sequential elsewhere; the same vote, round cap and BLOCK everywhere; the install command | [`harness/README.md`](../../harness/README.md), `scripts/install-review-process.sh` |
| VII | Incident, which layer, rule, bench case, scored against every case; `make bench` over `bench/corpus` | [`process/04-incident-evals.md`](../../process/04-incident-evals.md), [`process/00-seven-layers.md`](../../process/00-seven-layers.md) § 7 |
| VIII | What the CLI catches in Cloudflare config, with severity and action; what stays with the reviewer; the OWASP IDs | [`security/skills/cloudflare-config.md`](../../security/skills/cloudflare-config.md), [`internal/cli/testdata/golden/cloudflare-config.sample.txt`](../../internal/cli/testdata/golden/cloudflare-config.sample.txt) |
| card | The slip: `Decision: DO NOT MERGE`, `wrangler.toml:9`, `CRITICAL`, `BLOCK` | [`internal/cli/testdata/golden/cloudflare-config.sample.txt`](../../internal/cli/testdata/golden/cloudflare-config.sample.txt) |

## Use a plate

Use `<picture>` so the plate follows the reader's colour scheme. From the repository root:

```html
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/art/hero-dark.svg">
  <img src="docs/art/hero-light.svg" width="1280" alt="Engraved title plate: AI Code Review, seven layers of review for any agent harness. Seven slices of swiss cheese stand in a row; one defect is stopped at layer III, security that blocks, and another at layer V, risk-scaled human approval.">
</picture>
```

In `README.pt-BR.md`, point at `docs/art/pt-BR/` and use the Portuguese alt text. From a file under `process/`, the path is `../docs/art/`.

## Alt text

| Plate | Alt text |
|---|---|
| I | Engraved title plate: AI Code Review, seven layers of review for any agent harness. Seven slices of swiss cheese stand in a row; one defect is stopped at layer III, security that blocks, and another at layer V, risk-scaled human approval. |
| II | Engraved plate of the seven layers stacked as slabs, each with its role in this repo. A defect falls through holes in layers I and II and is stopped at layer III, security that blocks. |
| III | Engraved flow of eight stations: prepare, classify the tier, the security engine (which can stop the run with BLOCK), bug hunters in parallel, a verifier vote, a filter, the report, and a human who decides. |
| IV | Engraved plate of three pedestals for the risk tiers: low needs no human approval once AI review and security are green, medium needs one, high needs two. The change sets the tier, not the author. |
| V | Engraved plate on convergence: a spiral closes in over three rounds to "reconsider the approach", beside bar charts of the results the post's authors reported (65% to 78% passing after one round). |
| VI | Engraved orrery: the review process, defined once in reviewers and the skill, at the centre; claude-code, kiro, codex, cursor, copilot and gemini on the ring around it. |
| VII | Engraved wheel of five stations around make bench: an incident, which layer should have caught it, the answer becomes a rule, a bench case, and every change scored against every case. |
| VIII | Engraved gate with its portcullis down, stopping the diff. On the left, what the CLI catches with severity and action; on the right, what stays with the reviewer. |
| card | Engraved card: AI Code Review, seven layers of review for any agent harness. A report slip reads Decision: DO NOT MERGE and Action: BLOCK; a defect passes slices I and II and is stopped at III. |

Each SVG also carries `role="img"`, a `<title>` and a long `<desc>` that spells out every label on the plate and the source.

## Style

- **Engraving.** Every tone is a field of parallel lines drawn as an SVG `<pattern>`: five levels from light to dark, crossed at levels 4 and 5, at the angle that follows the form. Shadows are stipple, outlines are swelling strokes, and the paper is the only flat fill.
- **Palette.** Light: paper `#f4efe4`, ink `#1d1a16`, accent `#b3412c`. Dark: paper `#14130f`, ink `#e9e2d0`, accent `#e0735a`. The accent marks only what blocks.
- **Type.** EB Garamond, converted to outlines so the plates need no web font. Captions use faux small capitals; notes and figure lines are italic.
- **Layout.** A ruled frame with corner rosettes, the plate caption at the top, a figure line, and a source cartouche at the bottom. Single-layer plates add `Ref. §` and the post's heading, in English in both languages.
- **No marks.** Text labels only: no logos, wordmarks or brand colours. Anthropic and Cloudflare appear only as plain engraved text.

## Regenerate

The generator is a small build tool for these pictures only; the review engine itself is Go. It needs Node 18 or later, pinned dependencies in [`generator/package.json`](generator/package.json), and `rsvg-convert` to render `site/og.png`.

```bash
npm --prefix docs/art/generator ci
npm --prefix docs/art/generator run build
npm --prefix docs/art/generator run build -- --preview /tmp/plates   # also a PNG of every SVG
```

The build is deterministic: the same strings give the same bytes. Each SVG stays under 200 KB. The text lives in [`generator/src/strings/en.json`](generator/src/strings/en.json) and [`generator/src/strings/pt-BR.json`](generator/src/strings/pt-BR.json); the drawing code lives in `generator/src/pen`, `figures` and `plates`.

## Sources

- **Review process.** Addy Osmani, Morgan Westlee Lunt and Qing Wang, [*How the Claude Code team does code review*](https://claude.dev/blog/solving-code-review-at-anthropic/), Anthropic, 15 October 2026. https://claude.dev/blog/solving-code-review-at-anthropic/
- **Cloudflare (plate VIII).** developers.cloudflare.com: [Workers secrets](https://developers.cloudflare.com/workers/configuration/secrets/), [Access JWT validation](https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/) and [WAF](https://developers.cloudflare.com/waf/).
- **Typeface.** EB Garamond, Copyright 2017 The EB Garamond Project Authors (https://github.com/octaviopardo/EBGaramond12), licensed under the SIL Open Font License 1.1. The licence is in [`generator/fonts/OFL-EB-Garamond.txt`](generator/fonts/OFL-EB-Garamond.txt).

Plates drawn in code by Tiago Vilas Boas. Wording and drawings are this repo's own; mistakes are ours, not the post authors'.
