# Kiro adapter

Thin adapter over `reviewers/*.md`.

## Install

```sh
scripts/install-review-process.sh --harness kiro --target /path/to/repo
```

## What maps to what

| Review process | Kiro file |
|---|---|
| Always-on rules and report format | `.kiro/steering/review-process.md` (`inclusion: always`) |
| Role index (loaded on demand) | `.kiro/steering/review-roles.md` (`inclusion: manual`) |
| Pre-PR self check | `.kiro/hooks/pre-pr-self-check.json` (trigger `Stop`, agent action) |
| Run the review | `.kiro/steering/review-pr.md` (`inclusion: manual`): type `/review-pr` or `#review-pr` in chat |

## Limits and uncertainty

- Kiro has no documented way to spawn parallel, clean-context subagents from a steering file, so the roles run sequentially. Clear context between roles (new chat session).
- The files follow Kiro's docs as read on 2026-11-12: [steering](https://kiro.dev/docs/steering/) front matter (`inclusion`) and the [hooks](https://kiro.dev/docs/hooks/) file (`version`, `hooks[].trigger`, `hooks[].action`). They were checked against the docs, not run in Kiro. If the hook does not appear, create it from the Kiro Hooks panel and paste the `prompt` text.
- The review is a manual steering file, not a hook: Kiro's docs say new manual-trigger hooks cannot be created in the IDE and recommend a manually included steering file, which runs as a slash command.
- The hook is enabled on install. Set `"enabled": false` to silence it.
