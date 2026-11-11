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

## What the CLI checks and what stays with the reviewer
The Go CLI (`internal/rules/cloudflare*.go`, finding class `cloudflare-config`, secrets reported as
`secret`) catches these deterministically, on added lines only, with `path:line` and redacted
values. Test paths, placeholders, empty values, `${VAR}` references and Turnstile test keys stay silent.

| CLI rule (file) | Catches | Severity / action |
|---|---|---|
| wrangler vars (`cloudflare_wrangler.go`) | credential-looking value under `[vars]` / `"vars"` in `wrangler.toml`, `wrangler.json`, `wrangler.jsonc` | `CRITICAL` BLOCK (`secret`, CWE-798) |
| wrangler public URL (`cloudflare_wrangler.go`) | `workers_dev = true` or `preview_urls = true` added | `MEDIUM` REVIEW |
| public env secret (`cloudflare_public_env.go`) | Turnstile secret or `CLOUDFLARE_API_TOKEN` literal in a `NEXT_PUBLIC_*` / `VITE_*` / `PUBLIC_*` variable, or in client code | `CRITICAL` BLOCK (`secret`) |
| terraform settings (`cloudflare_terraform.go`) | `ssl = "off"`; `ssl = "flexible"`, `always_use_https = "off"`, `min_tls_version` 1.0 or 1.1, `security_level = "essentially_off"`, `browser_check = "off"`, `fight_mode = false` | `HIGH` BLOCK for `ssl = "off"`, else `MEDIUM` REVIEW |
| terraform rulesets (`cloudflare_terraform_ruleset.go`) | managed WAF ruleset disabled or skipped as a whole; rate limit rule disabled; rate limit ruleset removed (deleted `phase = "http_ratelimit"` line) | `MEDIUM` REVIEW |
| access header (`cloudflare_access.go`) | `Cf-Access-Authenticated-User-Email` read with no `Cf-Access-Jwt-Assertion` in the same file's diff | `HIGH`, REVIEW (validation may live elsewhere) |
| CORS (existing `cors` rule) | `Access-Control-Allow-Origin: *`, reflected origin with credentials in Worker code | `MEDIUM` REVIEW |

Limits of the CLI: it sees only the diff. A whole-file deletion has no new-file line, so a deleted
rate limit file is a reviewer item. Per-rule overrides (`enabled = false` inside `overrides`, a skip that
lists specific rules) are tuning and stay silent. `ssl = ...` needs a `cloudflare` mention in the diff or path.

**Reviewer-only** (the diff rarely proves them, so the CLI does not guess): Turnstile widget without a
server-side `siteverify`; admin route with no Access app or auth check; WAF custom rules (a
`cloudflare_ruleset` in the custom phase with a broad `skip`, or a block rule turned to `log`); rate
limit rules removed in a way that keeps the phase line; bot management beyond `fight_mode`
(`sbfm_*` levels lowered, JS detections off); AI Gateway bypassed; origin IP trust; the Access
and WAF state that lives in the dashboard.

## Instructions for the agent
1. Scope is the **provided PR diff only**: `wrangler.toml` / `wrangler.json(c)`, Worker and Pages
   Functions code, `_headers` / `_redirects`, and Terraform using the `cloudflare` provider.
2. Run the CLI first, from a trusted source and never from the checkout under review
   (`go run github.com/tiagovilasboas/ai-code-review/cmd/review@main <diff>`); its Cloudflare
   findings are evidence, do not re-derive them. Then look for these, **only if added or made worse in this diff** (rows the CLI
   already covers are listed above; report one only if the CLI missed it):

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
   | An AI Gateway config that drops an auth token (`cf-aig-authorization`), a rate limit, or caching/logging controls, or a gateway URL that stops being used by the model client | Anyone with the URL spends the provider key | `MEDIUM` |
   | A `cloudflare_ruleset` in the custom WAF phase (`http_request_firewall_custom`) removed, or a block rule changed to `log`/`skip` over a broad expression such as `true` or a whole path prefix | Custom WAF rule weakened | `MEDIUM` |
   | A rate limiting rule removed or its threshold raised sharply on a login, signup, password-reset or paid API route; or a login route added with no rate limit in the diff | Brute force and cost abuse | `MEDIUM` |
   | Bot Fight Mode / Super Bot Fight Mode / bot management levels lowered (`sbfm_definitely_automated = "allow"`, verified bots allowed on a sensitive path) | Automated traffic reaches forms and APIs | `MEDIUM` |
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
