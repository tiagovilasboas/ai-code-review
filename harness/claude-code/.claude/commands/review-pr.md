---
description: Run the multi-agent PR review (tier, security engine, parallel bug hunters, verifier votes). Usage - /review-pr [base-ref] [quick|standard|deep] [round]
argument-hint: "[base-ref] [quick|standard|deep] [round]"
---

Load `.agents/skills/pr-review-process/SKILL.md` and run its pipeline on the current branch.

Arguments: `$ARGUMENTS` (base ref default `main`, depth default from `REVIEW.md` else `standard`, round default 1).

Use the subagents in `.claude/agents/`: `risk-classifier`, `risk-second-opinion`, `security`, then `bug-hunter` in parallel (quick 1, standard 3, deep 5), then three `verifier` runs per candidate, and `convergence` from round 2. Each subagent gets fresh context: only the diff, the PR intent, `REVIEW.md` and `AGENTS.md`.

Report starts with `N findings (I important, N nit, P pre-existing)`. Do not edit files, merge or publish. Stop for a human decision.
