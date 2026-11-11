# Codex adapter

Thin adapter over `reviewers/*.md`.

## Install

```sh
scripts/install-review-process.sh --harness codex --target /path/to/repo
```

## What maps to what

| Review process | Codex file |
|---|---|
| Always-on rules | block in `AGENTS.md` between `review-process:begin/end` markers (appended if the file exists, never overwritten) |
| Orchestrator entry point | `.codex/prompts/review-pr.md` |

## Notes

- Codex reads custom prompts from `~/.codex/prompts/`. Copy or symlink `.codex/prompts/review-pr.md` there to get `/prompts:review-pr BASE=main DEPTH=deep`. Project-level prompt directories are not read by every Codex version.
- Codex has no declarative subagents: run roles as separate sessions, or let Codex run them sequentially with a context clear between roles.
- No hooks are shipped.
