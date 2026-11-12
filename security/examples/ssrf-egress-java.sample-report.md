# Sample report — SSRF / egress, Java (fixture)

Not a production incident. Walkthrough of [`security/skills/ssrf-egress.md`](../skills/ssrf-egress.md) against [`ssrf-egress-java.sample.diff`](ssrf-egress-java.sample.diff).

## Hunt
- Diff adds `proxy`: the `@RequestParam String url` reaches `RestTemplate.getForObject` (hunt table: caller-influenced URL).
- No scheme + host allowlist in the hunk.
- `billing` keeps a fixed host and only appends a path variable: **not** a finding.

## Findings

```text
HIGH | src/main/java/app/ProxyController.java:12 | caller-supplied @RequestParam url reaches RestTemplate.getForObject with no scheme+host allowlist | CWE-918
```

## Evidence
- Source: `src/main/java/app/ProxyController.java:11` — `@RequestParam String url`.
- Sink: `src/main/java/app/ProxyController.java:12` — `restTemplate.getForObject(url, ...)`.
- Explicit non-finding: `src/main/java/app/ProxyController.java:7` — fixed `https://billing.example.invalid` host; input is only a path segment.

## Verdict
**FAIL**

## Fix direction
Allowlist scheme + host (and port) before the call at `src/main/java/app/ProxyController.java:12`; disable redirects or re-validate the next hop. Human decides request-changes.

## Out of scope / not claimed
- No IMDS / VPC blast radius; no CVSS; no exploit PoC.
