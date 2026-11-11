# Gemini CLI adapter

Thin adapter over `reviewers/*.md`.

## Install

```sh
scripts/install-review-process.sh --harness gemini --target /path/to/repo
```

## What maps to what

| Review process | Gemini CLI file |
|---|---|
| Always-on rules | block in `GEMINI.md` between `review-process:begin/end` markers (appended if the file exists, never overwritten) |
| Orchestrator entry point | `.gemini/commands/review-pr.toml` (`/review-pr main deep`) |

## Notes

- Roles run sequentially in fresh sessions (`/clear` between roles); no declarative subagents or hooks are shipped.
- Gemini CLI can read `AGENTS.md` if you set `context.fileName` in its settings; otherwise rely on `GEMINI.md`.
