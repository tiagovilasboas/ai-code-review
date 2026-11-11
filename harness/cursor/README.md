# Cursor adapter

Thin adapter over `reviewers/*.md`.

## Install

```sh
scripts/install-review-process.sh --harness cursor --target /path/to/repo
```

## What maps to what

| Review process | Cursor file |
|---|---|
| Orchestrator rules | `.cursor/rules/review-process.mdc` (agent-requested: Cursor attaches it when the description matches) |
| Role index | `.cursor/rules/review-roles.mdc` (attaches when `reviewers/*.md` is in context) |

Cursor also reads `AGENTS.md` and `.agents/skills/`, so the `pr-review-process` skill works directly.

## Notes

- Fresh context = a new chat or a background agent per role. There are no declarative subagents; Cursor degrades to sequential roles.
- No slash command or hooks are shipped; type "run the review process on this branch".
