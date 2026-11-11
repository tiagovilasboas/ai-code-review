## Why

What problem does this change solve? Problem and constraints first, not a file list.

## What

What changed, and what is explicitly **out of scope**?

- [ ] Process (`process/`, `REVIEW.md`, `templates/`)
- [ ] Reviewer role or guardrail (`reviewers/`, `.agents/skills/pr-review-process/`)
- [ ] Security pass: skill, runbook or example (`security/`, `.agents/skills/appsec-pr-review/`)
- [ ] Engine, rules or catalog (`cmd/`, `internal/`, `bench/`)
- [ ] Harness adapter or installer (`harness/`, `scripts/install-review-process.*`)
- [ ] CI (`.github/`)
- [ ] Hygiene / docs only (LICENSE, CONTRIBUTING, `docs/`)

## Testing

Commands you ran and their **real output** (paste it; a summary does not count). The agent's claim is not evidence.

```bash
make check
```

## Risk

Tier from `go run ./cmd/tier --from origin/main --to HEAD` (low, medium or high) and why. Name any path that always needs a person.

## How to revert

One line: `git revert <sha>`, flip a flag, or what else undoes this.

## Prompter

If an agent wrote this PR: the person who directed it (counts as one required reviewer). Otherwise write "n/a".

## Review rounds

- [ ] This is round 1. After round 3 the reviewer asks you to reconsider the approach; do not keep patching line by line.

## Evidence contract

- [ ] Every claimed finding (skill/runbook PRs) requires `path:line` + one-line why
- [ ] No invented CWE / CVE / CVSS
- [ ] Agent still does **not** merge, deploy, or rotate secrets
- [ ] No exploit PoC or attack procedure

## Checklist

- [ ] English title and body
- [ ] Follows [CONTRIBUTING.md](../CONTRIBUTING.md)
- [ ] README layout table still accurate if a path was added
