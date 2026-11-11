# Claude Code adapter

Thin adapter. All role logic lives in `reviewers/*.md` (copied by the installer); these files only point at it.

## Install

```sh
scripts/install-review-process.sh --harness claude-code --target /path/to/repo
```

## What maps to what

| Review role | Claude Code file |
|---|---|
| bug-hunter, verifier, security, risk-classifier, risk-second-opinion, convergence | `.claude/agents/<role>.md` (read-only tools: Read, Grep, Glob, Bash) |
| pre-pr-self-check | `.claude/commands/pre-pr.md` (`/pre-pr`) |
| orchestrator (`pr-review-process` skill) | `.claude/commands/review-pr.md` (`/review-pr [base] [depth] [round]`) |

Each subagent runs in its own context window, which gives the clean context the process needs. Parallel hunters = several subagent calls in one turn.

## Optional hook

`hooks.example.json` (not installed) is a `Stop` hook that prints a reminder. Merge its `hooks` key into `.claude/settings.json` by hand if you want it.

## Notes

- No `model` is set in the agent files, so they inherit yours. Add one if you want cheaper hunters.
- The reviewers use Bash for read-only commands and the review engine; keep your permission rules strict.
