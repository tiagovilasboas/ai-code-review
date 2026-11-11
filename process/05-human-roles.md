# What humans still do

**English** · [Português (Brasil)](05-human-roles.pt-BR.md)

AI reviews the diff; people own the outcome. The point of the other layers is to spend human attention where it pays.

| Responsibility | Concretely |
|---|---|
| Decide whether to build | Before any code: is this worth doing? Review does not rescue a wrong idea |
| Review risky code | Billing, auth, auto-update, migrations, **CI workflows** (always, by rule) |
| Review architecture others will copy | A new pattern becomes the template for the next ten changes |
| Read evidence, not summaries | Test output, commands run, screenshots, the `path:line` in the finding. Not the agent's "all tests pass" |
| Own the outcome | You approved it, you own it |

## Blameless ownership
"You merged it, you own it" means you fix it and write the incident review. It does not mean you are blamed. The review asks **which layer should have caught this**, then changes the process: a lint rule, a skill line, an AGENTS.md/CLAUDE.md line, a bench case ([04](04-incident-evals.md)).

## For the prompter of an agent-written PR
You count as one required reviewer ([01](01-risk-tiers.md)). Before requesting others:
- Run the change yourself; do not forward an unrun diff.
- Read the pre-PR check output, not its conclusion.
- Be able to explain the approach without the agent.

## For reviewers
| Do | Don't |
|---|---|
| Open the evidence the report cites | Approve on the report's summary alone |
| Resolve a thread with a reason | Resolve to unblock |
| Say what you did not check | Imply full coverage |
| Approve the final commit | Rely on an approval of an earlier commit |

## Escalation
Security BLOCK, tier `high`, or round 3 without convergence: a person decides. See [`runbooks/00-overview.md`](../runbooks/00-overview.md) and [`guardrails/write-approval.md`](../guardrails/write-approval.md). The agent never merges.

Next: [06 adoption checklist](06-adoption-checklist.md).
