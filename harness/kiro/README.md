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
| Pre-PR self check | `.kiro/hooks/pre-pr-self-check.kiro.hook` (event `agentStop`) |
| Run the review | `.kiro/hooks/review-pr.kiro.hook` (manual trigger) |

## Limits and uncertainty

- Kiro has no documented way to spawn parallel, clean-context subagents from a steering file, so the roles run sequentially. Clear context between roles (new chat session).
- The steering front matter (`inclusion`) and the hook JSON shape (`when.type`, `then.type: askAgent`) follow Kiro's documented formats as best known; Kiro's schema may change. If a hook does not appear, create it from the Kiro Hooks panel and paste the `prompt` text.
- Hooks are enabled on install. Set `"enabled": false` to silence one.
