# Sample report — permissive CORS in framework config (fixture)

Not a production incident. Walkthrough of [`security/skills/secrets-config.md`](../skills/secrets-config.md) against [`cors-server-config.sample.diff`](cors-server-config.sample.diff). Shows Python and PHP config shapes (Go gin/rs-cors, Spring `CorsConfiguration` and Django `CORS_ALLOW_ALL_ORIGINS` follow the same shape).

## Findings
- `MEDIUM` | `app/main.py:7` | FastAPI `CORSMiddleware` allows every origin (`allow_origins=["*"]`) together with `allow_credentials=True` | CWE-942

## Evidence
- Sink: `app/main.py:7` — `allow_origins=["*"]`; credentials enabled at `app/main.py:8`.
- Not a finding: `config/cors.php:3` — an explicit origin list, even with `supports_credentials` at `config/cors.php:4`.

## Verdict
**REVIEW**

## Fix direction
List the exact front-end origins in `allow_origins`, or drop credentials.

## Out of scope / not claimed
- No exploit payload, no CVSS.
