# Claude Code adapter

Thin adapter. All role logic lives in `reviewers/*.md` (copied by the installer); these files only point at it.

## Install

```sh
scripts/install-review-process.sh --harness claude-code --target /path/to/repo
```

## What maps to what

| Review role | Claude Code file |
|---|---|
| bug-hunter, verifier, security, risk-classifier, risk-second-opinion, convergence | `.claude/agents/<role>.md` (bug-hunter, verifier, convergence, risk-second-opinion: Read, Grep, Glob; security and risk-classifier add Bash) |
| pre-pr-self-check | `.claude/commands/pre-pr.md` (`/pre-pr`) |
| orchestrator (`pr-review-process` skill) | `.claude/commands/review-pr.md` (`/review-pr [base] [depth] [round]`) |

Each subagent runs in its own context window, which gives the clean context the process needs. Parallel hunters = several subagent calls in one turn.

## Optional hook

`hooks.example.json` (not installed) is a `Stop` hook that prints a reminder. Merge its `hooks` key into `.claude/settings.json` by hand if you want it.

## Notes

- No `model` is set in the agent files, so they inherit yours. Add one if you want cheaper hunters.
- `tools:` in the agent files is enforced by Claude Code: bug-hunter, verifier, convergence and risk-second-opinion have no Bash. security and risk-classifier keep Bash because they run the engine and the tier classifier.
- Not enforced by the agent files: what those two run in Bash. The permission rules below narrow it; sandboxing bounds it.

## Reviewing fork PRs safely

Building or running code from a PR checkout executes the contributor's code on your machine, and lets the PR change the engine and the tier config that judge it. So the roles run the engine and classifier from a trusted source, never from the checkout:

- `go run github.com/tiagovilasboas/ai-code-review/cmd/review@main ...` and `.../cmd/tier@main ...`. The repo has no release tags yet; once one exists, replace `@main` with a tag or commit in the agent files and in `settings.example.json`.
- The tier config comes from the base branch: `git show origin/<base>:.github/review-tiers.json > <tmp file>`, then `--config <tmp file>`.
- The diff is read with `git diff --no-ext-diff --no-textconv <base>...HEAD`.

`settings.example.json` (not installed, like `hooks.example.json`) holds `permissions.allow` rules limited to those commands plus `git log`, `git show` and `git worktree add`, and `permissions.deny` rules for git's `--output` and `--ext-diff` options, building, testing, running scripts, network fetches, git config and git push. Merge its `permissions` key into `.claude/settings.json` (or `settings.local.json`) by hand. Deny is checked before allow, so `git diff --no-ext-diff --output=<file>` is refused even though it starts like an allowed command.

These rules narrow which commands run without asking; they are not a sandbox. Claude Code's own docs call Bash rules that constrain arguments fragile. Limits:

- Deny rules are an explicit list, not "everything else". Any Bash command not in `allow` falls back to your permission mode (a prompt by default), so never review a fork PR in a mode that auto-approves Bash.
- An allowed program still does what its other arguments say: `git show` can print any file in history.
- For a real boundary, run the review with Claude Code [sandboxing](https://code.claude.com/docs/en/sandboxing) on, or in a disposable container or VM.
