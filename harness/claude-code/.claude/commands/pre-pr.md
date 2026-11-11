---
description: Verify work before opening a PR. Runs checks with real output, then a fresh subagent reviews only the diff.
argument-hint: "[base-ref]"
---

Follow `reviewers/pre-pr-self-check.md` with base `$ARGUMENTS` (default `main`).

1. Run the repo's tests, type check, lint and build (see `AGENTS.md`/`Makefile`). Show the real output.
2. Write the PR intent in one paragraph: problem, approach, how tested, how to undo.
3. Spawn the `bug-hunter` subagent with ONLY the diff and that paragraph. Do not pass your own reasoning.
4. Fix real findings, re-run the checks. Maximum 3 rounds, then ask whether to reconsider the approach.

Do not open the PR or merge; report the checks and the review result.
