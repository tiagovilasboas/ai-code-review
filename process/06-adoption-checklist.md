# Adoption checklist

**English** · [Português (Brasil)](06-adoption-checklist.pt-BR.md)

## Before the PR
- [ ] Tests, build, lint run; output shown (not summarised)
- [ ] Fresh-context self-check of the diff ([`pre-pr-self-check`](../reviewers/pre-pr-self-check.md))
- [ ] PR description states intent, how it was tested, how to undo it

## Every PR
- [ ] Tier computed (`go run ./cmd/tier`) and posted with reasons
- [ ] Deterministic security engine ran; BLOCK resolved or explained
- [ ] AI review: hunters, then verifiers; findings labelled `Important | Nit | Pre-existing`
- [ ] Report starts with a count; round cap 3 respected
- [ ] Required CI green on every platform

## Approval
- [ ] Number of humans matches tier (0 / 1 / 2); CI workflow and `alwaysHuman` paths have a person
- [ ] Approver read the evidence, not the summary
- [ ] Approval is on the final commit; stale approvals dismissed
- [ ] Bot approvals not counted; prompter counted

## After merge
- [ ] Merge queue re-ran checks on `main`; flag on for internal first, staged release
- [ ] Revert path known; owner on call
- [ ] Incident? Add a bench case ([04](04-incident-evals.md)), name the layer that should have caught it
- [ ] Repeated finding? Fix at the source: lint rule, skill, or an AGENTS.md line. Repeated false positive? A `REVIEW.md` line

## Rollout: start in one repo, tune for weeks, widen
| Step | Do |
|---|---|
| 1. Install | `scripts/install-review-process.sh` into one repo; copy [`templates/REVIEW.md`](../templates/REVIEW.md) and [`templates/review-tiers.json`](../templates/review-tiers.json) |
| 2. Observe | Run reviewers in comment-only mode; no gating beyond the security engine |
| 3. Tune | Weekly: [`scripts/review-noise.sh`](../scripts/review-noise.sh), fix top false positives in `REVIEW.md` |
| 4. Gate | Turn on tier checks and required approvals once the clean-PR test shows no harmful reports |
| 5. Widen | Next repo, with that repo's own `REVIEW.md` and tiers. Never copy tiers blind |

## Reality check
The source process comes from a team reviewing its own widely known public code, close to the best case for an AI reviewer. Treat its numbers as an upper bound, not a forecast. Studies point the same way:

| Study | Finding |
|---|---|
| CRBugBench (arXiv 2610.13668) | Review agents found the bug within the first 5 comments 71% of the time on open-source repos, far lower on closed ones |
| TestSuiteBench (arXiv 2610.13065) | Agent-written test suites pass but catch under 60% of planted bugs |
| Follow-up fix study (arXiv 2609.26847) | Agent PRs need about 1.5x more follow-up fixes; agents fix about 70% of their own |

**Open vs closed:** your private code is not in any model's training data, so expect weaker review than the public benchmarks show. That is why the incident eval set ([04](04-incident-evals.md)) is built from **your** incidents, and why the deterministic engine, which does not depend on model familiarity, stays the blocking security gate. Measure yourself; do not take the claims on trust.
