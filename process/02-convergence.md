# Convergence: stop review that never ends

**English** · [Português (Brasil)](02-convergence.pt-BR.md)

**The problem** (numbers from the Claude Code team's post at Anthropic, [*How the Claude Code team does code review*](https://claude.dev/blog/solving-code-review-at-anthropic/)). One PR ran 50+ overnight review rounds. About 59% of what was flagged had already been there on the first pass: the reviewer kept finding old things in new light, and the author (an agent) kept acting on each. This repo is an independent implementation of that process, not affiliated with or endorsed by Anthropic.

## Controls
| Control | Rule |
|---|---|
| Round cap | After **3** rounds the reviewer stops listing and asks the author to reconsider the approach |
| Diff-only re-review | After the first pass, re-review only what changed since the last reviewed commit |
| Class of bug | Flag the class ("all handlers here skip the ownership check"), not one line at a time |
| No repeats | Never re-post a finding already posted, resolved, or answered with a reason |
| Nightly sweep | Existing code is reviewed by a scheduled run, not smuggled into unrelated PRs |

Implemented by [`reviewers/convergence.md`](../reviewers/convergence.md) and the orchestrating skill ([`pr-review-process`](../.agents/skills/pr-review-process/SKILL.md)). Pre-existing bugs are labelled `Pre-existing`, never `Important`, and never block the PR.

## Reported results at the source
| Metric | Before | After |
|---|---|---|
| PRs passing after one round | 65% | 78% |
| PRs needing 5+ rounds | 8.9% | 1.5% |
| Mean reviews per PR | 2.08 | 1.32 |

These are the originating team's numbers on their own repos. Measure yours ([03](03-noise-metrics.md)); do not assume them.

## Round log
Each review report ends with `Round N of 3`. Round 3 ends with the reconsider-the-approach note, listing the classes still open and nothing else. No round 4 without a human.

## Nightly sweep (zero-cost)
A scheduled CI job (or cron on any machine) runs `go run ./cmd/review --git --from <last-sweep> --to HEAD` plus the bug hunter on directories not touched recently. Findings go to an issue, not into an open PR.

## Failure modes
| Failure | Guard |
|---|---|
| Reviewer drifts into style | Style and cleanup-only findings are dropped ([`REVIEW.md`](../REVIEW.md) nit cap) |
| Author agent loops fixing nits | Nit cap, round cap, no-repeat rule |
| Cap hides a real bug | A security BLOCK is outside the cap; it stays until fixed or resolved with a reason |

Next: [03 noise and metrics](03-noise-metrics.md).
