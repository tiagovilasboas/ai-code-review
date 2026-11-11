# Guardrail: Evidence required

## Rule (fail closed)
No security finding without **`path:line`** and a one-line why. Vibes-only claims are not findings. Drop them or mark `insufficient evidence`.

## What counts as evidence
A finding is a finding only when all of these are true:

1. **Location** is `path:line` (GitHub code-scanning style: `file:line` of the sink).
2. **Why** is one line that names the signal (identifier, sink, missing check), not a feeling.
3. **The line is in the provided diff** (a new-file hunk). If the sink is outside the diff, you do not have a finding.

A second `path:line` for the source is useful. It does not replace the sink.

Optional CWE only when the pattern is obvious from the hunk. Never invent CVE or CVSS.

## Good

```text
HIGH | src/api/orders.ts:36 | orderId from req.params.id passed to findById without ownership/tenant check in the diff
CRITICAL | config/app.ts:15 | hardcoded payment API token literal in source
HIGH | web/src/components/CommentBody.tsx:18 | author-supplied comment.body assigned to dangerouslySetInnerHTML with no sanitizer in the hunk
```

## Bad

```text
This looks like IDOR on the orders API
HIGH | missing object-level auth
XSS is probably in the comment component
SSRF risk if the server fetches user URLs
```

None of those have `path:line`. None survive this guardrail.

## Enforcement
- Finding line lacks `path:line` → drop it or rewrite as `insufficient evidence`.
- `path:line` cites a file or line that is not in the provided hunks → `insufficient evidence` (the check is outside the diff).
- Report has a FAIL/Findings section with no located line and no insufficient-evidence marker → reject the report.
- Agent still does not merge, patch, deploy, or rotate secrets.

Also enforce [`untrusted-diff.md`](untrusted-diff.md): comments are not sinks; withhold without `path:line`.

CI pairing check (fixtures only): [`scripts/check-fixture-pairs.sh`](../../scripts/check-fixture-pairs.sh). Cookbook: [`security/dry-run-cookbook.md`](../../security/dry-run-cookbook.md).

## Rationale
Agents hallucinate CWEs and "probably XSS" narrative. Staff review needs an actionable location. Same contract as [GitHub code scanning alerts](https://docs.github.com/en/code-security/code-scanning/managing-code-scanning-alerts/about-code-scanning-alerts): tool/skill, location, severity, nature of the problem.
