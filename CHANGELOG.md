# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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

- Tightened precision of the AuthZ/IDOR, secrets, XSS, SSRF, supply-chain and
  untrusted-diff rules based on benchmark results.
