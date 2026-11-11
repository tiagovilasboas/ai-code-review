# The seven layers of code review

**English** · [Português (Brasil)](00-seven-layers.pt-BR.md)

No single check catches everything. Each layer has holes; stack them so the holes do not line up (the "swiss cheese" model). The question per change is not "who reviews this?" but **"what could go wrong with this change, and which check catches it best?"**

Guardrail metric for the whole process: **less human review must never mean more reverts or incidents.** Track reverts and fix-forwards per merged PR before and after every loosening. If they rise, tighten.

| # | Layer | Implemented here by | Zero-cost fallback |
|---|---|---|---|
| 1 | Verify before the PR | [`reviewers/pre-pr-self-check.md`](../reviewers/pre-pr-self-check.md) | Run lint/tests yourself, paste the output, re-read your own diff tomorrow |
| 2 | AI review on every PR | [`reviewers/bug-hunter.md`](../reviewers/bug-hunter.md), [`reviewers/verifier.md`](../reviewers/verifier.md), [`.agents/skills/pr-review-process/SKILL.md`](../.agents/skills/pr-review-process/SKILL.md), harness wiring in [`harness/`](../harness/README.md) | Run the skill in any harness you already pay for, or review with the checklist by hand |
| 3 | Security that blocks | [`reviewers/security.md`](../reviewers/security.md) plus the deterministic engine `cmd/review` ([`action.yml`](../action.yml), [`ai-review.yml`](../.github/workflows/ai-review.yml)) | The engine needs no LLM and no network |
| 4 | Blocking automated checks | Your CI; the engine is one required check | Any CI plus the PR template questions |
| 5 | Risk-scaled human approval | `cmd/tier`, [`reviewers/risk-classifier.md`](../reviewers/risk-classifier.md), [`review-tier.yml`](../.github/workflows/review-tier.yml), CODEOWNERS | `go run ./cmd/tier` is deterministic; no model needed |
| 6 | Merge queue and staged release | Your platform (merge queue, flags) | Require "branch up to date" before merge |
| 7 | Detect fast, roll back | Your flags and on-call | A documented revert command and an owner |

Evidence for incident learning lives in [`bench/`](../bench/README.md) (see [04](04-incident-evals.md)).

## 1. Verify before the PR exists
The author (human or agent) runs tests, build and lint and **shows the output**. Then a subagent with **fresh context** reviews only the diff: the author's reasoning contaminates its own review. Optional pre-PR passes: bug hunt, simplify, end-to-end check.
- Here: `reviewers/pre-pr-self-check.md`.
- Fallback: paste test output into the PR; a colleague (or a new session) reads the diff cold.

## 2. AI review on every PR
Read the diff against the whole repo **and the PR's stated intent**: does it solve the stated problem, fit the system, break something else? On bug fixes, look for the root cause, not the symptom. Several hunters work in parallel; separate **verifiers** vote on each finding; only findings enough verifiers agree on are posted. Style, diagnostics and cleanup-only findings are dropped. Labels: `Important`, `Nit`, `Pre-existing`. Depth per repo: `quick | standard | deep` (set in [`REVIEW.md`](../REVIEW.md)).
- Here: `bug-hunter` + `verifier` roles, orchestrated by the `pr-review-process` skill; convergence limits in [02](02-convergence.md).
- Fallback: one hunter pass plus one verifier pass in the same harness; drop anything the verifier cannot reproduce from the diff.

## 3. Security review that can block
A PR cannot merge while an open security or bug finding remains. Fix it, or resolve the thread **with a reason**; a later commit that fixes it clears it.
- Here: the deterministic engine blocks on exit code 1 (HIGH/CRITICAL, `path:line`, official IDs only; the one HIGH that stays REVIEW is the Cloudflare Access header rule, because the JWT check may live outside the diff). `reviewers/security.md` adds LLM judgement but cannot override a BLOCK.
- Fallback: the engine alone. It is the zero-cost, zero-drift security gate.

## 4. Automated blocking checks
Lint, types, tests on every platform, build. The PR template asks **how it was tested and how to undo it**. Required status checks, not advice.

## 5. Risk-scaled human approval
The change sets the tier, not the author. See [01](01-risk-tiers.md).

## 6. Merge queue and staged release
The queue re-runs checks against `main` with the change applied. Ship behind flags, internal first, then staged (latest before stable).

## 7. Fast detection and rollback
Flags turn off without a deploy; someone is on call. Every incident review asks **which layer should have caught it**, and the answer becomes a rule, a skill line, or a bench case ([04](04-incident-evals.md)).

Next: [01 risk tiers](01-risk-tiers.md).
