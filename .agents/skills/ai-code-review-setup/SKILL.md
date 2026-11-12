---
name: ai-code-review-setup
description: >-
  Installs the AI Code Review multi-agent review process into the user's repository
  through the coding agent. Use when the user asks to install, set up or add AI Code
  Review (for example "instala o AI Code Review neste repositório"). Detect the
  harness, show a dry-run plan, write only after a clear yes, then verify. Never
  install toolchains silently. A human approves; do not merge or publish.
license: MIT
compatibility: Claude Code, Codex, Cursor, Kiro, GitHub Copilot, Gemini CLI. Needs git, a POSIX shell (or PowerShell on Windows) and Go for the engine.
metadata:
  audience: installer
---

# AI Code Review setup

The deterministic Go engine is the core; you run it for the user. Do this in order and
stop at any failed step.

1. **Find the installer.** If `scripts/install-review-process.sh` is not in the current
   directory, clone `https://github.com/tiagovilasboas/ai-code-review` (shallow) into a
   temporary directory outside the target repo and use its `scripts/`. The target is the
   user's repository (the current directory unless told otherwise).
2. **Detect harnesses** in the target: `.claude/` or `CLAUDE.md` (claude-code), `.kiro/`
   (kiro), `.codex/` (codex), `.cursor/` (cursor), `.github/copilot-instructions.md`
   (copilot), `.gemini/` or `GEMINI.md` (gemini). If none is found, install for the
   harness you are running in; if unsure, ask. Use `all` only when the user asks.
3. **Plan.** Run `scripts/install-review-process.sh --harness <name|all> --target <dir> --dry-run`
   (on Windows, `scripts/install-review-process.ps1`). Show the plan. Mention
   `--with-security-kit` (adds `security/skills/` and the `appsec-pr-review` skill) as
   optional. `conflict` lines mean an existing file differs; never use `--force` unless
   the user says so.
4. **Ask** for a clear yes. Without it, write nothing.
5. **Install.** Re-run the same command without `--dry-run`. Exit 3 means conflicts were
   kept as they were; report them.
6. **Verify.**
   - The review command file for the harness exists (`.claude/commands/review-pr.md`,
     `.codex/prompts/review-pr.md`, `.gemini/commands/review-pr.toml`,
     `.github/prompts/review-pr.prompt.md`, `.kiro/steering/review-pr.md`, or
     `.cursor/rules/review-process.mdc` for Cursor).
   - The engine runs: `review --version`, or
     `go run github.com/tiagovilasboas/ai-code-review/cmd/review@main --version`.
     If Go is missing, say so and point to
     `https://tiagovilasboas.github.io/ai-code-review/docs/quickstart.html`. Never
     install Go or any other toolchain without the user's explicit yes.
7. **Report** what was created, how to run the review next (`/review-pr` in Claude Code,
   Gemini, Copilot and Kiro; `/prompts:review-pr` in Codex after copying the prompt to
   `~/.codex/prompts/`; in Cursor, ask "run the review process on this branch"), and that `REVIEW.md` and `.github/review-tiers.json` are the
   files to edit. Do not commit or publish anything unless asked.
