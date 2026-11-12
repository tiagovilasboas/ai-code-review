# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Portuguese (Brazil) is now the default language of the human-facing docs.
  `README.md`, `harness/README.md`, `docs/*.md` and `process/*.md` are in
  Portuguese; each English version moved to `*.en.md` and is linked from the
  top of the page. Links across the repo, the site and `llms.txt` point at the
  English pages where the text around them is English.

### Added

- Six borderless engraved vignettes for the README sections (why deterministic,
  quickstart, what it catches, install, how it works, guardrails), in Portuguese
  and English, light and dark. They have no frame and no background, and are
  drawn with new `BARE` themes in the art generator.
- `scripts/test-install-review-process.ps1` and a CI job that runs it with
  `pwsh` on Windows and Linux. On a machine with `sh` it also checks that both
  installers write the same file list.
- The 7-layer code review process the Claude Code team at Anthropic published
  ("How the Claude Code team does code review", 2026-10-15, https://claude.dev/blog/solving-code-review-at-anthropic/): `process/` docs for the
  layers, risk tiers, convergence, noise metrics, incident evals, human roles
  and an adoption checklist. This repo is an independent implementation, not
  affiliated with or endorsed by Anthropic.
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
- Engraved plates for the README and the site (`docs/art/`, English and pt-BR,
  light and dark), each crediting the Anthropic post, and the generator that
  builds them (`docs/art/generator/`, an art-only Node tool).

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
  `security/skills/injection.md` and OWASP catalog entries.
- Precision/recall benchmark harness (`cmd/bench`, `make bench`) over a labelled
  corpus of diffs, with held-out cases for XSS, SSRF, supply-chain and
  untrusted-diff rules.
- Local gates: versioned lefthook hooks and `make check` (gofmt, vet, tests,
  cross-OS vet, fixture pairing).
- CodeQL workflow for Go and the GitHub Actions workflows; workflow actions pinned to full SHAs.
- Release tooling: GoReleaser configuration, a tag-triggered release workflow
  with build provenance attestation, and `make release-check`.

### Changed

- Kiro adapter follows Kiro's current docs: the pre-PR hook is
  `.kiro/hooks/pre-pr-self-check.json` (`trigger: Stop`, agent action), and the
  review runs from a manual steering file, `.kiro/steering/review-pr.md`
  (`/review-pr`), instead of a legacy manual hook. A repo that installed the old
  `*.kiro.hook` files keeps them; delete them by hand.
- Copilot prompt file uses `agent: agent`; `mode` is the old name of the field.
- The README (both languages), `CITATION.cff` and `llms.txt` now lead with the
  seven-layer review process; the security engine is presented as its layer 3.
- The layout follows the review process. The security pass (layer 3) lives in
  `security/`: `skills/`, `examples/`, `runbook.md` (was
  `runbooks/00-overview.md`) and `dry-run-cookbook.md`. The guardrails moved to
  `reviewers/guardrails/`, next to the roles that follow them, so the installer
  always copies them; the golden reports moved to `internal/cli/testdata/golden/`.
  SARIF help links now point to `security/skills/`.
- `internal/diff` records removed hunk lines, so rules can see a protection
  being turned off.
- Cloudflare Turnstile test secrets are treated as documentation placeholders,
  not as leaked secrets.
- Tightened precision of the AuthZ/IDOR, secrets, XSS, SSRF, supply-chain and
  untrusted-diff rules based on benchmark results.

### Fixed

- `untrusted-diff` no longer blocks ordinary docs. A bare `approve` or `merge`
  counts only when aimed at an AI addressee (after a vocative such as
  `AI reviewers:`, or right next to the name), and a name inside a path, file
  name or URL (`.claude/`, `CLAUDE.md`, `ai-code-review`) is not an addressee.
  A docs line that only quotes an injection phrase is REVIEW instead of BLOCK.
  The XSS, SSRF, AuthZ and injection rules stay silent on prose documents
  (`.md`, `.txt`, `.rst`, `.adoc`; not `.mdx`). On this repo's own docs that
  takes 28 blocking findings to 0, with 6 quoted mentions left as REVIEW. The
  corpus gains four cases (329); the former blind set is unchanged at 47/6/9.
- `internal/diff` no longer reads an added line that starts with `++ ` as a
  new file header. The hunk's line counts decide where it ends, so such a
  line can no longer hide the added lines after it from the rules.

### Security

- `ai-review.yml`: Claude gets no shell and no `gh`. Trusted steps write the
  diff, the PR text and the commit log to files outside the checkout, and a
  trusted step posts Claude's structured output after redacting token-shaped
  strings. The run uses `--restricted`, `--strict-mcp-config`, user settings
  only, an env scrub, and denies reads of `/proc`, the runner temp dir and
  `.git/`. Claude's start-up files (`.claude/`, `CLAUDE.md`, `.mcp.json` and
  similar) are reset to the base, the review instructions come from a base
  worktree, and the previous reviewed SHA is taken only from this workflow's own
  marker comment and must be in the head's history. The header no longer
  implies it blocks merges.
- `review-tier.yml`: the `human-approvals` job is documented as an advisory
  signal, because on `pull_request_review` GitHub runs the workflow from the PR
  merge ref; enforcement is native branch protection (required approvals, Code
  Owners, dismiss stale approvals). It counts only reviewers with write access,
  always runs, and fails closed when classification was cancelled or failed
  (except the bootstrap case). Classifier outputs are allow-listed, the
  base-branch job no longer persists credentials or saves a Go cache, and paths
  can no longer close the fenced block in the PR comment.
- The classifier and the diff parser (`cmd/tier`, `internal/tier`,
  `internal/diff`) are high risk with a code owner. More paths need a person
  and a code owner: `scripts/action/**`, the tier config, `CODEOWNERS`, `Makefile`, the installer test, `go.mod`, `go.sum`,
  `REVIEW.md` and `.agents/skills/**`.
- Reviewer roles run the engine and the classifier from a trusted source (the
  upstream default branch or a clean base-branch worktree), never from the
  checkout under review, and read the tier config from the base branch. The
  Claude Code bug-hunter, verifier and convergence agents no longer get Bash,
  and `harness/claude-code/settings.example.json` ships matching permission
  rules.
- The Action's job summary escapes paths, titles and the decision reason from
  the diff, so a crafted path cannot break the table or inject HTML.
- `ci.yml` runs with a read-only token, and its "comment-only sinks are
  withheld" check now fails the job when a CWE-79 or CWE-639 finding appears
  (the old negated `grep` never failed under `set -e`). The release workflow
  runs `make check` in a separate read-only job before the job that can
  publish.

### Removed

- The Node CLI (`bin/review.js`, `src/`, `test/`, `package.json`). The Go CLI
  printed byte-identical reports on the examples both engines covered and is now the only engine:
  run `go run ./cmd/review <diff>` or `go install …/cmd/review@latest` instead of
  `npm run review`. CI and the release workflow no longer set up Node.
- `.github/dependabot.yml`.
