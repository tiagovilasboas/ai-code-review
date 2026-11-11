---
description: Run the multi-agent PR review on the current branch
argument-hint: BASE=main DEPTH=standard
---

Load `.agents/skills/pr-review-process/SKILL.md` and run its pipeline on the current branch.

Base ref: $BASE (default `main`). Depth: $DEPTH (default from `REVIEW.md`, else `standard`).

Run each role from `reviewers/` in fresh context: risk classifier, risk second opinion, security engine, bug hunters (quick 1, standard 3, deep 5), then verifiers (3 per candidate). If you cannot spawn sub-agents, run the roles sequentially and clear context between them. Report starts with `N findings (I important, N nit, P pre-existing)`. Do not edit files or merge; stop for a human decision.
