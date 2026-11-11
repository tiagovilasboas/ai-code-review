# Role — Pre-PR self check

## Purpose
Verify work before a PR exists. The author agent proves it with real output; a fresh subagent then reviews only the diff, free of the author's reasoning.

## Inputs
- Author agent: its working tree and the repo's test, build and lint commands (from `AGENTS.md`, `Makefile` or CI config).
- Fresh reviewer: ONLY the diff (`git diff <base>...HEAD`), the PR intent in one paragraph, `REVIEW.md` and `AGENTS.md`.

## Instructions
Author agent:
1. Run the repo's tests, type check, lint and build. Show the real command output (the tail is fine), not a summary. If a command was not run, say so.
2. Do not claim "done" on failing or skipped checks.
3. Write the PR intent as one paragraph: problem, approach, how it was tested, how to undo it.
4. Hand the diff and intent to a fresh subagent (a new subagent, or a new session if the harness has none). Do not pass your reasoning or logs of attempts.
5. Optional extra passes: bug hunt (`reviewers/bug-hunter.md`), simplify, one end-to-end check.

Fresh reviewer: follow `reviewers/bug-hunter.md` on the diff alone.

6. Fix real findings and re-run the checks. Then open the PR.

## Output format
`Checks:` each command with exit code and the key output line. `Review:` the bug-hunter report (`N findings ...`). `Intent:` the paragraph.

## Stop conditions
Checks green with output shown, and a fresh review with no open Important finding. Cap the self-review loop at 3 rounds (`reviewers/convergence.md`).
