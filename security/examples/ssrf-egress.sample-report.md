# Sample report — SSRF / egress (fixture)

Not a production incident. Walkthrough of [`security/skills/ssrf-egress.md`](../skills/ssrf-egress.md) against [`ssrf-egress.sample.diff`](ssrf-egress.sample.diff).

## Hunt
- Diff adds `previewLink`: `req.body.url` is passed to server `fetch` (hunt table: caller-influenced URL).
- Filter is a `localhost` / `127.0.0.1` string denylist — skill: denylist-only is FAIL.
- Client options set `redirect: "follow"`. Raw `response.text()` is returned as `{ html }`.
- `pingStatus` fetches a compile-time `https://status.example.invalid/health` — anti-pattern table: **not** a finding.

## Findings

```text
HIGH | src/api/preview.ts:14 | caller-supplied req.body.url reaches fetch with no scheme+host allowlist; denylist-only at :11; redirects follow | CWE-918
MEDIUM | src/api/preview.ts:16 | raw fetched body returned to the client
```

## Evidence
- Source: `src/api/preview.ts:9` — `targetUrl = req.body.url`.
- Denylist: `src/api/preview.ts:11` — `includes("localhost")` / `includes("127.0.0.1")`. Not an allowlist of scheme + host.
- Sink: `src/api/preview.ts:14` — `fetch(targetUrl, { redirect: "follow" })`.
- Response leak: `src/api/preview.ts:16` — `res.json({ html })` of `response.text()`.
- Explicit non-finding: `src/api/preview.ts:4` — hardcoded `https://status.example.invalid/health`. Input is not the URL.
- Explicit withhold: no metadata host, VPC, or IMDS path in the hunk → **insufficient evidence** for “IMDS is exposed”. Do not invent that finding.

## Verdict
**FAIL**

## Fix direction
Allowlist scheme + host (and port) before `fetch` at `src/api/preview.ts:14`. Drop the localhost string denylist as the only control. Disable redirects, or re-validate the next hop against the same allowlist. Do not return the raw fetched body. Network egress deny-by-default is ops follow-up, not a line you proved here.

## Out of scope / not claimed
- No IMDS / cloud-metadata blast radius. You cannot see VPC egress from this PR.
- No CVSS, no “this is in production”, no exploit PoC, no scheme-smuggling payload.
- `listPreviews` is unchanged → not a finding.
- Agent does not patch the handler. Human decides request-changes.
