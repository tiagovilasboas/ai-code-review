# Runbook: Dry-run cookbook

## Purpose
Run **one** skill on a fixture, compare the output to the expected report, and reject vibes-only claims. Time box: one pass.

This is not a real PR review and not an agent runtime. Human still decides merge vs request changes.

## Pairing

Index: [`security/examples/README.md`](examples/README.md). Guardrail: [`evidence-required`](../reviewers/guardrails/evidence-required.md). CI enforces the pairs via [`scripts/check-fixture-pairs.sh`](../scripts/check-fixture-pairs.sh).

| Diff | Expected report | Skill |
|---|---|---|
| [`security/examples/sample-pr.diff`](examples/sample-pr.diff) | [`security/examples/sample-report.md`](examples/sample-report.md) | `authz-idor` then `secrets-config` (two passes) |
| [`security/examples/xss-sink.sample.diff`](examples/xss-sink.sample.diff) | [`security/examples/xss-sink.sample-report.md`](examples/xss-sink.sample-report.md) | `xss-html` |
| [`security/examples/ssrf-egress.sample.diff`](examples/ssrf-egress.sample.diff) | [`security/examples/ssrf-egress.sample-report.md`](examples/ssrf-egress.sample-report.md) | `ssrf-egress` |
| [`security/examples/supply-chain.sample.diff`](examples/supply-chain.sample.diff) | [`security/examples/supply-chain.sample-report.md`](examples/supply-chain.sample-report.md) | `supply-chain` |
| [`security/examples/untrusted-comment.sample.diff`](examples/untrusted-comment.sample.diff) | [`security/examples/untrusted-comment.sample-report.md`](examples/untrusted-comment.sample-report.md) | `untrusted-diff` (always-on guardrail) |

Naming: `name.sample.diff` pairs with `name.sample-report.md`. Stage 1 exception: `sample-pr.diff` pairs with `sample-report.md`.

## How to run a pass

1. Pick **one** row. Do not load the whole skill pack.
2. Read the skill file (`security/skills/<name>.md`) and the guardrail.
3. Give the agent only that skill plus the `.diff` (paste or path). Scope is the diff, not imagined repo context.
4. Collect lines shaped like GitHub code-scanning alerts:

```text
SEVERITY | file:line | why
```

5. Compare to the paired `*-report.md`. Severity and wording may differ; the **cited `path:line` must land in a hunk of that diff**.
6. Drop any line that lacks `path:line`. Mark `insufficient evidence` when the sink is outside the diff.
7. Stop. Human records request-changes vs proceed. The agent does not merge, patch, or rotate secrets.

```bash
# From the repo root: confirm fixtures still pair (same check as CI)
bash scripts/check-fixture-pairs.sh
```

## What good evidence looks like

Required on every finding: **`path:line` of the sink** and a one-line why. A second `path:line` for the source is better, not a substitute.

Good (from the AuthZ fixture):

```text
HIGH | src/api/orders.ts:36 | orderId from req.params.id passed to findById without ownership/tenant check in the diff
```

Good (explicit non-finding, still located):

```text
insufficient evidence | src/api/preview.ts:4 | hardcoded status URL; input is not the URL; do not invent IMDS
```

Staff bar:

- The path exists in the paired diff (`+++ b/path` / hunk).
- The line number falls inside a new-file hunk (not a guessed line in unseen code).
- Why names the signal from the skill hunt table (identifier, sink, missing check), not a vibe.

## What to reject

Reject the line (or the whole report) when any of these are true:

| Reject | Why |
|---|---|
| `This looks like IDOR on the orders API` | No `path:line` |
| `HIGH \| missing object-level auth` | Severity without a location |
| `XSS in the comment component` | Component name is not evidence |
| Invented CWE / CVE / CVSS | Not obvious from the diff |
| Exploit PoC, payload, or curl attack | Out of scope for this kit |
| Finding whose `path:line` is outside the provided hunks | Check is outside the diff |
| Mixing several skills into one blob | Humans cannot audit the pass |

Silence is correct:

```text
insufficient evidence
no evidence-based findings
```

Do not upgrade those to a finding to look thorough.

## CI contract

`scripts/check-fixture-pairs.sh` fails the job when:

1. An `security/examples/*.diff` has no paired report (or a `*report.md` has no paired diff).
2. A finding line (`SEVERITY | ...`) lacks a `path:line` token.
3. A cited `path:line` is not inside a hunk of the paired diff.
4. The report has neither a finding with `path:line` nor an explicit `insufficient evidence` / `no evidence-based findings` marker.

The script does not run a model. It only checks pairing and the evidence pattern.

## Done when

- One skill, one fixture, one report (or two named passes on `sample-pr.diff`)
- Every finding has `path:line` inside a hunk, or the report says insufficient evidence
- No PoC, no merge, no invented catalog IDs
- Local `bash scripts/check-fixture-pairs.sh` exits 0
