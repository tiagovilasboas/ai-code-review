# Contributing

Thanks for helping keep this kit small, evidence-first, and usable in a 15-minute PR review.

This repository is an **actionable kit** for agentic pull-request review:

| Path | Role |
|---|---|
| `skills/` | What the agent checks |
| `.agents/skills/appsec-pr-review/SKILL.md` | Loadable Agent Skill (Cursor / Claude). Complementary to `npm run review` |
| `runbooks/` | When a human must step in |
| `guardrails/` | What must never be skipped (`evidence-required`, `untrusted-diff`, `write-approval`) |
| `examples/` | Sample PR diff + expected report (dry-run only) |
| `scripts/` | Fixture pairing check (CI) |

It is **not** a scanner SaaS, a dump of 50 prompts, or an “AI review” that invents findings.

## Quality bar

Every contribution must survive a Staff AppSec review. Align with how practitioners and platforms already report risk:

- **OWASP IDs, cited not invented** — map the skill to ASVS 5.0 and/or OWASP Top 10 / API / Agentic Top 10 / Agentic Skills Top 10 IDs that the sink actually exercises. Put app IDs in the skill table and in the CLI catalog (`src/owasp-catalog.js`). Put ASI/AST IDs under reviewer posture. Do not paste a full Top 10. Do not invent CWE / CVE / CVSS / catalog numbers.
- **OWASP GenAI / agentic framing** — agents take actions. Skills inspect diffs; humans decide merge. No skill grants the agent merge, deploy, or secret-rotation authority (`AST03`, `ASI09:2026`).
- **Practitioner checklists** — prefer a short, fail-closed checklist over narrative. Every step must be runnable on a PR diff.
- **Evidence path (GitHub code-security style)** — a finding is not a finding without a location. Report `path:line` (or `file:line`), a one-line why, and severity. Same contract as [code scanning alerts](https://docs.github.com/en/code-security/code-scanning/managing-code-scanning-alerts/about-code-scanning-alerts): tool/skill name, location, severity, nature of the problem.

Hard rules:

1. **Evidence or silence.** No `path:line` → drop the finding or mark `insufficient evidence`. See [`guardrails/evidence-required.md`](guardrails/evidence-required.md).
2. **Do not invent CWE / CVE / CVSS.** If the mapping is not obvious from the diff, omit it.
3. **Fail closed on agency.** The agent never merges. Human-in-the-loop is the default (overreliance and excessive agency are first-class risks).
4. **English** for skills, runbooks, guardrails, issues, and PRs. Paths, commands, and identifiers stay in English inside fenced blocks.

Severity labels stay consistent with GitHub security severity: `CRITICAL` / `HIGH` / `MEDIUM` / `LOW`. If unsure, use `LOW` or say insufficient evidence.

## What belongs here

A skill, runbook, or guardrail lands only if all of these are true:

1. **It is a reusable review contract**, not a one-off prompt for a single repo.
2. **It is harness-agnostic** — works when a human pastes a diff into any agent. No vendor lock-in, no required MCP, no required CI product.
3. **It was (or can be) run on a real PR diff** and produces either evidence-backed findings or an explicit “no evidence-based findings”.
4. **It names the human gate** — when to stop, escalate, or refuse.

What does **not** belong:

- Unverified claims, blog-only threat names, or “this looks like XSS” without a location
- Exploit proof-of-concepts, payloads, or attack procedures
- Skills that tell the agent to apply patches, rotate secrets, or merge
- Vendor-specific playbooks that only work inside one IDE or one scanner
- Prompt dumps without Purpose / Instructions / Output format
- An eval harness, an awesome list, or a runtime / layer model — this repo is an AppSec PR-review mechanism (CLI + skills + ASVS)

## Adding a skill

Skills live in `skills/`. Filename is English kebab-case.

```bash
touch skills/my-skill.md
```

Use this structure (match existing files such as `skills/authz-idor.md`):

```markdown
# Skill — Short title

## Purpose
One sentence: what risk this skill looks for in a PR diff.

## Risk class
Table of official IDs only (OWASP Top 10 / API / ASVS 5.0 for the app sink; ASI/AST for reviewer posture). Not the full Top 10.

## Instructions for the agent
1. Numbered, fail-closed steps.
2. Scope is the **diff**, not the whole repo, unless the runbook says otherwise.
3. Every finding: `file:line` + one-line why.
4. If unsure: **insufficient evidence** — do not invent CWE.

## Output format
- `SEVERITY` | `file:line` | summary | (optional CWE only if obvious)
```

Acceptance for a skill PR:

- [ ] Purpose is a single risk, not “review the whole PR”
- [ ] Instructions are numbered and runnable on a diff
- [ ] Output format requires `file:line` (GitHub alert-style location)
- [ ] Uncertain cases are `insufficient evidence`, not guessed CWEs
- [ ] No merge / exploit / payload steps
- [ ] English only; commands and paths in fenced blocks

Propose new skills with the **Add a skill** issue template before a large write-up.

## Adding a runbook

Runbooks live in `runbooks/`. They describe **human + agent** flow, not extra model instructions.

```bash
touch runbooks/nn-short-name.md
```

Minimum sections:

```markdown
# Runbook — Title

## Purpose
When this flow is used (time box if possible).

## Flow
1. **Human** provides the PR diff (or points the agent at the PR).
2. **Agent** runs one skill at a time.
3. **Guardrail:** every finding needs `path:line`.
4. **HITL:** human decides merge / request changes — agent does not merge.

## Done when
- Findings list **or** explicit "no evidence-based findings"
- No finding without path:line
```

Number new runbooks after `00-overview.md` (`01-…`, `02-…`). Do not replace the 15-minute overview unless you are fixing it.

## Adding a guardrail

Guardrails are fail-closed invariants. Keep them short.

```bash
touch guardrails/my-guardrail.md
```

```markdown
# Guardrail — Title

## Rule (fail closed)
The invariant in one sentence.

## Rationale
Why Staff review would reject a report that skips this.

## Enforcement
What to drop or mark when the rule is broken.
```

New guardrails must not weaken [`guardrails/evidence-required.md`](guardrails/evidence-required.md).

## How to run a review (local)

Clone, run the CLI, or load the Agent Skill and one markdown skill on a diff.

```bash
git clone https://github.com/tiagovilasboas/ai-code-review.git
cd ai-code-review
npm run review -- examples/xss-sink.sample.diff
```

Then:

```bash
# 1. Read the human + agent flow
#    runbooks/00-overview.md
#
# 2. Optional dry-run. Cookbook: runbooks/01-dry-run-cookbook.md
#    Index: examples/README.md
#
# 3. CLI (deterministic, no LLM) or Agent Skill + one markdown skill
#    npm run review -- <diff>
#    .agents/skills/appsec-pr-review/SKILL.md
#    skills/authz-idor.md
#    skills/secrets-config.md
#    skills/xss-html.md
#    skills/ssrf-egress.md
#    skills/supply-chain.md
#    skills/injection.md   (Go CLI only: bin/review-go <diff>)
#
# 4. Enforce the evidence guardrail before you trust the report
#    guardrails/evidence-required.md
#
# 5. Confirm fixture pairs and CLI tests (same checks as CI)
#    bash scripts/check-fixture-pairs.sh
#    npm test
```

A finding line should look like a code-scanning annotation, not a paragraph:

```text
HIGH | src/api/orders.ts:142 | object id from path used without ownership check
```

If the agent cannot cite a line:

```text
insufficient evidence: identifier usage is outside the provided diff
```

## Fixture pairing (CI)

Every `examples/*.diff` must have a paired report. CI runs `scripts/check-fixture-pairs.sh` on pull requests.

Naming:

- `name.sample.diff` pairs with `name.sample-report.md`
- `sample-pr.diff` pairs with `sample-report.md` (Stage 1 names)

The check fails when:

- a diff has no report (or a `*report.md` has no paired diff)
- a finding line (`SEVERITY | ...`) lacks `path:line`
- a cited `path:line` is not inside a hunk of the paired diff
- the report has neither finding evidence nor an explicit `insufficient evidence` / `no evidence-based findings` marker

How to run a fixture pass: [`runbooks/01-dry-run-cookbook.md`](runbooks/01-dry-run-cookbook.md).

```bash
bash scripts/check-fixture-pairs.sh
```

## Language and style

- **English** for all new contributor-facing text (this file, issue/PR templates, skills, runbooks, guardrails).
- Human-facing docs (`README.md`, `process/`, `harness/README.md`, `docs/review-tier.md`, `docs/ai-review-workflow.md`) also have a Brazilian Portuguese mirror (`*.pt-BR.md`). English is the source of truth: when you change one of them, update its `.pt-BR.md` in the same PR or say in the PR body that the mirror is behind. Agent-facing prompts (`reviewers/`, `skills/`, `SKILL.md`, `REVIEW.md`) stay English only.
- Commands, paths, and identifiers stay in **English fenced blocks**. Do not mix another language into the same fence.
- Prefer checklists and tables over long prose.
- Cite a public source when you introduce a risk name (OWASP Top 10, ASVS 5.0, Agentic Top 10, Agentic Skills Top 10, GitHub code-security docs). Accuracy matters more than coverage.

## Pull request checklist

- [ ] English title and body; fill every template section: Why, What, Testing (the commands and their output), Risk, How to revert, Prompter (for agent-written PRs) and Review rounds
- [ ] Only the paths that belong to the change (`skills/`, `.agents/skills/`, `runbooks/`, `guardrails/`, `examples/`, `scripts/`, `src/`, or hygiene docs)
- [ ] New skill/runbook follows the templates above
- [ ] Findings contract still requires `path:line`
- [ ] New or changed fixtures still pair (diff + report) and pass `bash scripts/check-fixture-pairs.sh`
- [ ] No exploit PoC, no new vendor lock-in, no agent-merge instructions
- [ ] README Purpose / Run it / OWASP refs / Limit still accurate if you added a path or ID
- [ ] Changed `scripts/install-review-process.*` or `harness/`? `sh scripts/test-install-review-process.sh` passes (it runs in `make check`)
- [ ] `go run ./cmd/tier --from origin/main --to HEAD` shows the tier you expect; the PR gets that many human approvals

## Review process

PRs are reviewed as Staff AppSec artifacts: problem first, evidence second, agency last. Expect questions of the form “where is the line?” and “why does the agent stop here?”.

This repo uses its own 7-layer process: [`REVIEW.md`](REVIEW.md) sets what reviewers check here (depth, skip paths, nit cap 3, round cap 3), `review-tier.yml` labels each PR with its risk tier, and [`process/`](process/00-seven-layers.md) explains the layers.

Stage 1 added denser skills (`authz-idor`, `secrets-config`), a denser overview runbook, and `examples/` (`sample-pr.diff`, `sample-report.md`) for dry-runs. Stage 2 added `xss-html`, `ssrf-egress`, `supply-chain`, and `examples/xss-sink.sample.*`. Stage 3 densified `supply-chain` (anti-pattern table) and added `examples/supply-chain.sample.*` plus `examples/README.md`. Stage 6 added `examples/ssrf-egress.sample.*` so every skill has a fixture. Stage 7 added fixture-pairing CI (`scripts/check-fixture-pairs.sh`), the dry-run cookbook (`runbooks/01-dry-run-cookbook.md`), and a thicker evidence guardrail (good/bad examples, hunk-scoped `path:line`). The fixture index is [`examples/README.md`](examples/README.md) (one row per skill). Prefer link polish over new skills unless a clear hole remains. Keep examples free of exploit PoCs.

## Local gates

Run once after cloning to install the hooks (lefthook, pinned in `go.mod`):

```sh
make hooks
```

- pre-commit: gofmt on staged `*.go` files and `go vet ./...`.
- pre-push: `make check`.

`make check` mirrors CI: gofmt, `go vet`, `go test`, cross-OS vet (windows and linux), `npm test`, and the fixture pairing check. `make build` writes `bin/review-go`.
