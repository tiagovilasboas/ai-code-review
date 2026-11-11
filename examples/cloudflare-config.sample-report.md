# Sample report — Cloudflare config (fixture)

Not a production incident. Walkthrough of [`skills/cloudflare-config.md`](../skills/cloudflare-config.md) against [`cloudflare-config.sample.diff`](cloudflare-config.sample.diff). The Go CLI reports all four findings on its own; a reviewer pass adds what is outside the diff.

## Hunt
- `wrangler.toml` adds `PAYMENTS_API_KEY` under `[vars]` with a literal value (hunt table: credential under vars).
- `wrangler.toml` adds `workers_dev = true`: the `*.workers.dev` URL answers outside the zone WAF and Access.
- `infra/zone.tf` adds `ssl = "flexible"`: traffic from Cloudflare to the origin is in clear text.
- `src/admin.ts` reads `Cf-Access-Authenticated-User-Email` with no `Cf-Access-Jwt-Assertion` check in the file.
- `TURNSTILE_SITE_KEY` is a public site key and `min_tls_version = "1.2"` is the hardened value → **not** findings.

## Findings

```text
CRITICAL | wrangler.toml:9 | literal payments API key under [vars]; it ships in the bundle, move it to wrangler secret | CWE-798
HIGH | src/admin.ts:3 | Access identity header trusted with no Cf-Access-Jwt-Assertion validation in this file | CWE-290
MEDIUM | infra/zone.tf:4 | ssl mode set to flexible (origin leg in clear text) | CWE-319
MEDIUM | wrangler.toml:4 | workers_dev enabled; the workers.dev URL bypasses zone WAF, Access and rate limits | CWE-288
```

## Evidence
- Secret: `wrangler.toml:9` — `PAYMENTS_API_KEY` under the `[vars]` table opened at `wrangler.toml:6`. Value withheld on purpose.
- Public URL: `wrangler.toml:4` — `workers_dev = true`.
- TLS: `infra/zone.tf:4` — `ssl = "flexible"`.
- Access: `src/admin.ts:3` — header read; no JWT validation visible. The HIGH is REVIEW, not BLOCK: validation may live in another file.
- Explicit non-finding: `wrangler.toml:8` — Turnstile site key, public by design.
- Explicit non-finding: `infra/zone.tf:5` — TLS 1.2 minimum.

## Verdict
**FAIL**

## Fix direction
Move the payments key to `wrangler secret put` and rotate it, because it is now in git history. Set `workers_dev = false` unless the Worker is meant to be public on that hostname. Use `ssl = "strict"`. Verify the `Cf-Access-Jwt-Assertion` JWT (signature, audience, issuer) before using the identity header.

## Out of scope / not claimed
- Whether an Access application is configured in the dashboard for `src/admin.ts` is outside the diff → **insufficient evidence**.
- No call to the Cloudflare API, no `wrangler`, no rotation of the key. A human does that.
