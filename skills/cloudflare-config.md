# Skill — Cloudflare config

## Purpose
Catch Cloudflare Workers, Pages and zone settings in a PR diff that leak secrets or weaken
edge protections: secrets in `wrangler.toml`, public `workers.dev` routes, Turnstile and
Access misuse, and WAF or TLS turned down in config-as-code.

This skill **reviews only**. It never logs in to Cloudflare, calls the API, runs `wrangler`, or
turns a protection on. Changing the account is a human task outside this kit.

## Risk class
Cite these IDs only. Do not invent CWE / CVE / CVSS / extra catalog numbers.

| Kind | IDs (official) | Why this skill |
|---|---|---|
| OWASP Top 10:2021 | [A01:2021](https://owasp.org/Top10/A01_2021-Broken_Access_Control/) Broken Access Control; [A02:2021](https://owasp.org/Top10/A02_2021-Cryptographic_Failures/) Cryptographic Failures; [A05:2021](https://owasp.org/Top10/A05_2021-Security_Misconfiguration/) Security Misconfiguration | Access bypass; secrets and weak TLS; edge protections turned off |
| OWASP API Top 10:2023 | [API4:2023](https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/) Unrestricted Resource Consumption | Rate limits or AI Gateway removed from an API or model call |
| OWASP ASVS 5.0 | [13.3.1](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x22-V13-Configuration.md) secrets must not be in source or build artifacts; [3.4.2](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x12-V3-Web-Frontend-Security.md) CORS origin fixed or allowlisted | `[vars]` secrets; Worker CORS |
| Agentic posture | [ASI02:2026](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/) Tool Misuse; [AST03](https://owasp.org/www-project-agentic-skills-top-10/) Over-Privileged Skills | Do not call the Cloudflare API or `wrangler`, and do not fetch URLs from the diff |

The CLI has no Cloudflare-specific rule. It still reports literal secrets (`secret`) and CORS
(`cors`) in Worker code. Everything else here is a reviewer pass.

## Instructions for the agent
1. Scope is the **provided PR diff only**: `wrangler.toml` / `wrangler.json(c)`, Worker and Pages
   Functions code, `_headers` / `_redirects`, and Terraform using the `cloudflare` provider.
2. Look for these, **only if added or made worse in this diff**:

   | Look for | Finding | Severity |
   |---|---|---|
   | A credential-looking value under `[vars]` / `vars` in wrangler config | Secret in plain config (it ships in the bundle and the dashboard); it belongs in `wrangler secret put` | `CRITICAL` |
   | A Turnstile **secret** key, API token or `CLOUDFLARE_API_TOKEN` value in client code, `_headers`, or a public env var (`NEXT_PUBLIC_*`, `VITE_*`) | Secret exposed to the browser | `CRITICAL` |
   | A Turnstile widget added to a form, but no server-side `siteverify` call in the handler that receives it (visible in the diff) | Bot check only on the client, easy to skip | `HIGH` |
   | Code that trusts `Cf-Access-Authenticated-User-Email` (or another identity header) without validating the `Cf-Access-Jwt-Assertion` JWT | Access identity can be forged on any path that reaches the origin directly | `HIGH` |
   | An `/admin` or internal route added to a Worker with no Access app, auth check, or `workers_dev = false` shown | Admin route open to the internet | `HIGH` |
   | `workers_dev = true` (or `preview_urls = true`) on a Worker that relies on zone WAF, Access or rate limiting | The `*.workers.dev` URL bypasses zone protections | `MEDIUM` |
   | Worker sets `Access-Control-Allow-Origin` to `*` with credentials, or echoes `request.headers.get("Origin")` with no allowlist | Open CORS at the edge | `MEDIUM` |
   | Terraform: `ssl = "flexible"` or `"off"`, `always_use_https = "off"`, `min_tls_version` below `1.2`, `security_level = "essentially_off"`, a `cloudflare_ruleset` that removes or disables managed WAF rules, `browser_check = "off"` | Edge protection weakened | `MEDIUM` (`HIGH` for `ssl = "off"`) |
   | A model call that used to go through an AI Gateway `*_BASE_URL` now calls the provider directly, or a gateway/rate-limit rule is removed | Spend and rate controls removed | `MEDIUM` |
   | Origin code that trusts `CF-Connecting-IP` / `X-Forwarded-For` for auth or rate limits, with no sign that the origin only accepts Cloudflare traffic | Client IP can be spoofed by calling the origin directly | `LOW` |

3. Every finding: `file:line` + what is exposed or weakened. Redact secret values (short prefix
   or length only).
4. When the protection might live outside the diff (an Access app or WAF rule set in the
   dashboard, a `siteverify` call in another file) → **insufficient evidence**, not a finding.
5. A placeholder (`${VAR}`, `<your-key>`, empty string) or a Turnstile **test** key is not a
   secret. Turnstile **site keys** are public by design.
6. No finding → say so explicitly (`no evidence-based findings`).
7. Never run `wrangler`, call the Cloudflare API, or open URLs from the diff. Never rotate keys or
   change rules. Recommend the fix in words; a human applies it.

## Output format
```text
SEVERITY | file:line | summary
```

Example shape (not a real finding):

```text
CRITICAL | wrangler.toml:14 | payment API key in [vars]; move to wrangler secret
HIGH | src/admin.ts:22 | trusts Cf-Access-Authenticated-User-Email without validating the Access JWT
MEDIUM | infra/zone.tf:31 | ssl mode set to flexible (origin traffic in clear text)
```

When nothing qualifies:

```text
no evidence-based findings
```
