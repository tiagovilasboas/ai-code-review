# Noise and metrics

[Português (Brasil)](03-noise-metrics.md) · **English**

Priority order: **catch real bugs first, then cut rounds.** Noise costs more with agent authors, because an agent acts on every nit and the PR churns.

## Classify every author response
| Class | Meaning | What to do |
|---|---|---|
| fixed | Author changed code because of the finding | True positive; keep |
| said wrong | Author showed the finding is incorrect | False positive; add a "stop flagging" line to `REVIEW.md` if it repeats |
| intentional | Behaviour is deliberate | Document the intent in AGENTS.md or a code comment |
| didn't matter | Correct but irrelevant | Candidate nit; raise the severity bar |

## Clean-PR test
Run the reviewer on PRs known to be clean (already shipped, no later fix-forward or revert). Count reports that would lead to a **harmful change** if an agent applied them. Target: zero; any hit is a rule to tighten.

## Compute
[`scripts/review-noise.sh`](../scripts/review-noise.sh) tallies the four classes, rounds per PR and findings per PR from review threads. See the script header for inputs; it is read-only and needs no LLM.

## Target signals
| Signal | Healthy direction |
|---|---|
| Share of findings classed `fixed` | Up |
| `said wrong` + `didn't matter` share | Down |
| Passes after one round | Up |
| PRs with 5+ rounds | Down |
| Mean reviews per PR | Down, but never at the cost of the next row |
| Reverts and fix-forwards per merged PR | **Flat or down.** If up, tighten |

## Failure modes
| Failure | Guard |
|---|---|
| Optimising rounds, losing bugs | Track bugs caught first; the incident eval ([04](04-incident-evals.en.md)) is the gate |
| Counting nits as wins | Only `fixed` counts as signal |
| Metrics gamed by closing threads | Resolve requires a reason; sample resolutions monthly |

Next: [04 incident evals](04-incident-evals.en.md).
