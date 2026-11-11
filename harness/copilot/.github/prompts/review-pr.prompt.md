---
description: Run the multi-agent PR review on the current branch
mode: agent
---

Load `.agents/skills/pr-review-process/SKILL.md` and run its pipeline on the current branch against `${input:base:main}`, at the depth set in `REVIEW.md` (default standard).

Run each role from `reviewers/` in a fresh context, sequentially if you cannot spawn agents, and clear context between roles. Report starts with `N findings (I important, N nit, P pre-existing)`. Do not edit files or merge; stop for a human decision.
