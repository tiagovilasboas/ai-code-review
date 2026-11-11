# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- The 7-layer code review process the Claude Code team at Anthropic published
  ("How the Claude Code team does code review", 2026-10-15, https://claude.dev/blog/solving-code-review-at-anthropic/): `process/` docs for the
  layers, risk tiers, convergence, noise metrics, incident evals, human roles
  and an adoption checklist.
- `REVIEW.md` for this repo (depth, skip paths, nit cap 3, round cap 3,
  count-first summary) and copyable `templates/REVIEW.md` and
  `templates/review-tiers.json` for adopters.
- Harness-neutral reviewer roles in `reviewers/` (pre-PR self-check, bug hunter,
  verifier, security, risk classifier, risk second opinion, convergence) and the
  orchestrator skill `.agents/skills/pr-review-process/SKILL.md`.
- Harness adapters in `harness/` for Claude Code, Kiro, Codex, Cursor, GitHub
  Copilot and Gemini CLI, plus `scripts/install-review-process.sh` (and `.ps1`)
  that installs them into another repo, with a test run by `make check`.
- Risk-tier classifier `cmd/tier` (`make tier`): reads a diff or a git range and
  prints the tier, the human approvals it needs and the reasons, as text or JSON.
  Configured by `.github/review-tiers.json`; CI workflow files always need a human.
  `"approvals": "advisory"` makes the approval check warn instead of fail, for
  single-maintainer repos such as this one.
- Workflows: `review-tier.yml` labels each PR `risk:*` / `needs-human` and counts
  human approvals on the head commit; `ai-review.yml` runs an opt-in multi-agent
  review capped at 3 rounds; `review-evals.yml` runs the bench corpus nightly
  and fails on any regression against `bench/results.json`. `scripts/review-noise.sh` tallies author responses to findings.
- `cloudflare-config` finding class with Go CLI rules for `wrangler.toml` /
  `wrangler.json(c)`, Worker code and Terraform `cloudflare` resources, 21 bench
  cases (`cf-01` to `cf-21`), and a sample diff with its report.
- `.github/CODEOWNERS` and PR template sections for testing, risk, how to revert,
  the prompter and review rounds.

- Go CLI (`cmd/review`, standard library only) that reviews a unified diff and
  fails closed: a HIGH AuthZ/IDOR or CRITICAL secret with `path:line` and an
  ASVS 5.0 ID is BLOCK.
- JSON and SARIF 2.1.0 report output (`--format`, `--output`), plus `--version`
  (set at link time through `internal/version.Version`).
- Diff input from a file, standard input or a repository: `--git` reads the
  workspace changes, a `--from`/`--to` range or a single `--commit`.
- Composite GitHub Action (`action.yml`) with a job summary and optional SARIF
  upload to Code Scanning, and a self-test workflow for it.
- Injection rules for JavaScript/TypeScript, Python, Go, Java and PHP: SQL
  injection, OS command injection and path traversal, with the matching
  `skills/injection.md` and OWASP catalog entries.
- Precision/recall benchmark harness (`cmd/bench`, `make bench`) over a labelled
  corpus of diffs, with held-out cases for XSS, SSRF, supply-chain and
  untrusted-diff rules.
- Local gates: versioned lefthook hooks and `make check` (gofmt, vet, tests,
  cross-OS vet, fixture pairing).
- CodeQL workflow for Go and JavaScript; workflow actions pinned to full SHAs.
- Release tooling: GoReleaser configuration, a tag-triggered release workflow
  with build provenance attestation, and `make release-check`.

### Changed

- `internal/diff` records removed hunk lines, so rules can see a protection
  being turned off.
- Cloudflare Turnstile test secrets are treated as documentation placeholders,
  not as leaked secrets.
- Tightened precision of the AuthZ/IDOR, secrets, XSS, SSRF, supply-chain and
  untrusted-diff rules based on benchmark results.
