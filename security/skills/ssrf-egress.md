# Skill — SSRF / egress

## Purpose
Find the server fetching a URL or host the caller can influence — webhook, preview, import, proxy, “fetch this image”. Diff review, not a network pentest.

## Risk class
Cite these IDs only. Do not invent CWE / CVE / CVSS / extra catalog numbers.

| Kind | IDs (official) | Why this skill |
|---|---|---|
| OWASP Top 10:2021 | [A10:2021](https://owasp.org/Top10/2021/A10_2021-Server-Side_Request_Forgery_%28SSRF%29/) Server-Side Request Forgery | Caller-influenced URL/host reaches a server client |
| OWASP ASVS 5.0 | [1.3.6](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x10-V1-Encoding-and-Sanitization.md) SSRF allowlist of protocols, domains, paths, ports; [13.2.4](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x22-V13-Configuration.md) outbound allowlist; [15.3.2](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x24-V15-Secure-Coding-and-Architecture.md) do not follow redirects unless intended | Allowlist + redirect posture in the hunk |
| Agentic posture | [ASI02:2026](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/) Tool Misuse and Exploitation; [AST03](https://owasp.org/www-project-agentic-skills-top-10/) Over-Privileged Skills | The reviewer must not `fetch` the URL in the diff. Location + why is enough. |

CLI prints `A10:2021, ASVS-5.0-1.3.6` when the rule fires. ASI/AST stay here.

## Hunt table

| Signal in the diff | What you are looking for | Evidence to cite |
|---|---|---|
| `fetch` / `axios` / `got` / `http.get` / `request` / `curl` in server code | URL, hostname, or path from query/body/header/DB | client call `path:line` + where the URL is built |
| Webhook / callback / “notify URL” / RSS / oEmbed / link-preview | User-supplied destination, often HTTPS-looking | handler `path:line` |
| Import / avatar / PDF / screenshot / “unfurl” | Server-side GET of an attacker URL | fetch site `path:line` |
| `file://`, `gopher://`, `dict://`, `ftp://`, or scheme from input | Parser + scheme confusion | URL parse `path:line` |
| Redirects left on (`maxRedirects`, `followRedirect: true`) | Allowlist on the first hop, attacker on the second | client options `path:line` |
| Cloud metadata hosts (`169.254.169.254`, `metadata.google.internal`) reachable from the same client | Egress to IMDS — only if the code can be pointed there | destination `path:line` |
| Denylist of `localhost` / `127.0.0.1` / RFC1918 only | Bypass via DNS rebinding, IPv6, decimal IP, redirects | filter `path:line` |

CLI language coverage: JS/TS (any non-literal URL), plus Python (`requests`, `httpx`, `urllib`), Go (`http.Get`/`NewRequest`, `client.Get`, resty `R().Get`), Java (`RestTemplate`, `WebClient.uri`, `new URL`/`URI.create`, Apache `HttpGet`) and PHP (`file_get_contents`, `curl_init`/`CURLOPT_URL`, Laravel `Http::`, Guzzle clients). In those four the destination must trace to request input; a fixed host, configured base URL or allowlist check earlier in the hunk is not flagged.

SSRF is not HTTP-only. If the diff opens `file://` or a non-HTTP scheme from input, that is in scope.

## Anti-patterns (look-alikes)

Hunt is the signal. This table is the trap: do not promote a look-alike to IMDS, and do not drop a real open-fetch FAIL.

| In the hunk | Usually **not** a finding | Usually **FAIL** (cite `path:line`) |
|---|---|---|
| Hardcoded `fetch("https://status.example/health")` | Fixed host. Input is not the URL | URL/host from query, body, header, or DB |
| `if (url.includes("localhost"))` denylist | — | Denylist-only. Cite the fetch line |
| “this could hit IMDS” | No controllable URL in the diff → `insufficient evidence` | Controllable URL with no scheme+host allowlist. Cite the fetch. Do not invent metadata theft |

## Required evidence
Every finding: **`path:line`** of the outbound call (or the URL constructor that feeds it).

If the URL is hardcoded to a trusted host and input is only a path on that host → say so; that is usually not FAIL unless the path can hit a dangerous internal route *and* you can cite it.

No `path:line` → drop it. Guardrail: [`reviewers/guardrails/evidence-required.md`](../../reviewers/guardrails/evidence-required.md).

## Pass / fail
- **FAIL** — user-influenced URL/host reaches a server HTTP(S) or scheme-capable client; no positive allowlist of scheme + host (+ port). Denylist-only is FAIL.
- **PASS** — destination is an allowlist (or a fixed internal client), redirects disabled or also checked, and the hunk does not pass raw responses of the fetched URL back to the browser.
- **INSUFFICIENT** — “this client *could* be pointed at IMDS” without a controllable URL in the diff. Do not invent CWE-918.

You cannot see VPC egress from a PR. Do not claim “network layer is open” or “IMDS is exposed in prod”. You can only talk about the code path.

## Fix direction
1. Allowlist scheme + host (and port). Do not denylist `localhost` and call it done — [OWASP A10:2021](https://owasp.org/Top10/2021/A10_2021-Server-Side_Request_Forgery_%28SSRF%29/) is explicit on this.
2. Disable HTTP redirects, or re-validate the next hop against the same allowlist.
3. Do not return the raw fetched body to the client.
4. After DNS resolve, reject loopback / link-local / private / metadata if the feature does not need them. Parser mismatches are real; prefer a library that does not do the “decimal IP” surprise.
5. Network egress deny-by-default is the real backstop — note it as ops follow-up, not as a finding you proved in the diff.

## Out of scope
- Writing exploit payloads, metadata token theft, or “hit this IMDS path”.
- Claiming blast radius in AWS/GCP/Azure without infra in the diff.
- XXE→SSRF unless the XML parser hunk is actually in the PR (then cite it; otherwise another skill).
- Full DNS-rebinding lab. If you cannot cite the fetch line, stop.

## Output format
- `SEVERITY` | `file:line` | summary | (optional CWE)
- One line: who controls the URL, whether redirects/allowlist are in the hunk.

## Refs (public)
- [OWASP Top 10:2021 A10 SSRF](https://owasp.org/Top10/2021/A10_2021-Server-Side_Request_Forgery_%28SSRF%29/)
- [OWASP ASVS 5.0 `1.3.6`](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x10-V1-Encoding-and-Sanitization.md)
- [OWASP SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [CWE-918](https://cwe.mitre.org/data/definitions/918.html) — only when the fetch line is in the hunk

Worked fixture: [`security/examples/ssrf-egress.sample.diff`](../examples/ssrf-egress.sample.diff) → [`security/examples/ssrf-egress.sample-report.md`](../examples/ssrf-egress.sample-report.md). Index: [`security/examples/README.md`](../examples/README.md).
