# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| `main` | Yes |
| Anything else (tags, forks, local copies) | No. Reproduce on `main` first |

This project is a deterministic PR-diff review kit. Its security posture depends on two things: the review rules must not let a BLOCK-class issue through, and the review must not be steered by the code it reviews.

## Reporting a vulnerability

**Do not open a public issue for a security problem.**

Report privately through GitHub Security Advisories:

<https://github.com/tiagovilasboas/ai-code-review/security/advisories/new>

Include, when you can:

- The smallest diff hunk that reproduces the problem (remove secrets first).
- The command you ran (for example `go run ./cmd/review <diff>`) and the output you expected versus the output you got.
- The commit SHA you tested.

Do not include live credentials, exploit payloads for production systems, or personal data.

## Response time

Best effort, within 7 days, to acknowledge the report. Fix timing depends on severity and on whether a rule change needs a fixture pair (see [CONTRIBUTING.md](CONTRIBUTING.md)).

## Scope

### In scope

- **False negatives that let a BLOCK-class issue through.** A HIGH AuthZ/IDOR or CRITICAL secret finding with `path:line` and an ASVS 5.0 ID that the CLI or a documented skill should have reported and did not.
- **Prompt injection that changes the decision.** Text inside a diff, comment, file name, or commit message that causes the agent or the CLI to change its verdict, suppress a finding, or claim approval.
- **Supply chain of this repository.** Compromised dependencies, workflows, install scripts, or release artifacts that this repository ships or runs.

### Out of scope

- **False positives** and wording issues. Report them with the [false-positive template](.github/ISSUE_TEMPLATE/false-positive.yml) as a public issue.
- **Findings in other projects** reviewed with this kit. Report those to the owners of those projects.
- **Vulnerabilities in third-party tools** (the Go toolchain, GitHub, LLM providers) that do not come from this repository's code. Report those upstream.
- **Social engineering, denial of service against GitHub infrastructure, and physical attacks.**
- **Theoretical issues** without a reproducing diff or command.

## Disclosure

Please give the maintainer a reasonable window to ship a fix before public disclosure. Credit is given in the advisory unless you ask otherwise.
