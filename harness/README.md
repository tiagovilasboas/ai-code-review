# Harness adapters

**English** · [Português (Brasil)](README.pt-BR.md)

The review process is defined once, in `reviewers/*.md` (roles) and `.agents/skills/pr-review-process/SKILL.md` (orchestrator). Each folder here is a thin adapter: it only points the harness at those files. No role logic is duplicated.

Install with `scripts/install-review-process.sh --harness <name|all> --target <repo>` (or `.ps1` on Windows). See each folder's README.

## Capability matrix

| Harness | Subagents | Parallel hunters | Slash command | Hooks | Result |
|---|---|---|---|---|---|
| claude-code | yes (`.claude/agents`) | yes | `/review-pr`, `/pre-pr` | yes (optional example) | full process |
| kiro | no declarative | no | manual hook trigger | yes (`agentStop`, manual) | sequential, context cleared between roles |
| codex | no declarative | no (separate sessions possible) | custom prompt (`~/.codex/prompts`) | no | sequential |
| cursor | no declarative (background agents) | manual | no | no | sequential |
| copilot | no declarative | no | prompt file `/review-pr` | no | sequential |
| gemini | no declarative | no | `/review-pr` | no | sequential |

## How it degrades

When a harness cannot run fresh, parallel subagents, the orchestrator runs the roles one after another and clears context between them (new session, or carry over only the diff, the PR intent and the candidate list). The process, the voting rule (2 of 3; deep 3 of 3), the round cap of 3 and the security BLOCK stay the same. Only independence is weaker, and the report says the review ran sequentially.
