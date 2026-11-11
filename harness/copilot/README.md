# GitHub Copilot adapter

Thin adapter over `reviewers/*.md`.

## Install

```sh
scripts/install-review-process.sh --harness copilot --target /path/to/repo
```

## What maps to what

| Review process | Copilot file |
|---|---|
| Always-on rules | block in `.github/copilot-instructions.md` between `review-process:begin/end` markers (appended if the file exists, never overwritten) |
| Orchestrator entry point | `.github/prompts/review-pr.prompt.md` (run it from Copilot Chat in VS Code, `/review-pr`) |

## Notes

- Prompt files are a VS Code / JetBrains Copilot Chat feature. `mode: agent` may be named `agent: agent` in newer versions; edit the front matter if your version asks.
- Copilot has no declarative subagents here: roles run sequentially, new chat between roles.
- Copilot's own PR review reads `.github/copilot-instructions.md`, so the rules there also steer it.
