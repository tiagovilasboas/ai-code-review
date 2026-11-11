---
inclusion: manual
---

# Review roles

Each role is a prompt in `reviewers/`. Run one role per fresh chat session (or fresh subagent) and give it only the diff, the PR intent, `REVIEW.md` and `AGENTS.md`.

| Role | File |
|---|---|
| Bug hunter | `reviewers/bug-hunter.md` |
| Verifier (one candidate per run) | `reviewers/verifier.md` |
| Security (can block) | `reviewers/security.md` |
| Risk classifier | `reviewers/risk-classifier.md` |
| Risk second opinion | `reviewers/risk-second-opinion.md` |
| Pre-PR self check | `reviewers/pre-pr-self-check.md` |
| Convergence | `reviewers/convergence.md` |

Kiro runs one agent per session, so run roles sequentially and clear context between them (new session, or carry over only the diff and the candidate list). Say in the report that the review was sequential.
