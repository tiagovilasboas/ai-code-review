# Skill — Secrets & config

## Purpose
Catch hardcoded secrets and dangerous config defaults introduced or worsened in a PR diff.

## Risk class
Cite these IDs only. Do not invent CWE / CVE / CVSS / extra catalog numbers.

| Kind | IDs (official) | Why this skill |
|---|---|---|
| OWASP Top 10:2021 | [A02:2021](https://owasp.org/Top10/A02_2021-Cryptographic_Failures/) Cryptographic Failures; [A05:2021](https://owasp.org/Top10/A05_2021-Security_Misconfiguration/) Security Misconfiguration | Secrets in source; debug / open CORS / insecure defaults |
| OWASP ASVS 5.0 | [13.3.1](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x22-V13-Configuration.md) secrets must not be in source or build artifacts; [3.4.2](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x12-V3-Web-Frontend-Security.md) CORS `Access-Control-Allow-Origin` is fixed or allowlisted | Hardcoded credential vs `origin: *` |
| Agentic posture | [ASI09:2026](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/) Human-Agent Trust Exploitation; [AST03](https://owasp.org/www-project-agentic-skills-top-10/) Over-Privileged Skills | Report location only. Never echo the full secret. Never rotate, delete, or write. |

CLI prints `Class: Secret in source`, `ASVS-5.0-13.3.1`, `Action: BLOCK`, and `Decision: DO NOT MERGE` for a hardcoded credential. CORS `*` is `REVIEW` (`A05:2021`, `ASVS-5.0-3.4.2`). ASI/AST stay here.

## Instructions for the agent
1. Scope is the **provided PR diff only**.
2. Look for newly added or changed literals that look like credentials: API keys, tokens, private URLs with embedded userinfo, `password =`, `secret =`, PEM blocks, cloud access keys.
3. Flag debug flags, `CORS` allow-all (`*`), disabled TLS verification, or default admin credentials **only if** introduced or made worse in this diff.
4. Every finding: `file:line` + what class of secret/config leaked (e.g. "hardcoded API token", "CORS origin *"). Redact the secret value in the summary (show at most a short prefix or length).
5. If a string might be a placeholder/test fixture with no production path in the diff → **insufficient evidence** or `LOW` with that caveat — do not invent CVEs.
6. No finding → say so explicitly (`no evidence-based findings`).
7. Never write exploit PoCs. Never instruct the agent to rotate, commit `.env`, or push secrets. Human rotates credentials outside this kit.

## Output format
```text
SEVERITY | file:line | summary | (optional CWE only if obvious)
```

Example shape (not a real finding):

```text
CRITICAL | config/app.ts:18 | hardcoded payment API token in source
MEDIUM | src/server.ts:9 | CORS origin set to * in production bootstrap path
```

When nothing qualifies:

```text
no evidence-based findings
```

Severity: `CRITICAL` for live-looking credentials in source; `HIGH`/`MEDIUM` for misconfiguration that widens exposure; `LOW` when the context is unclear.

Worked fixture: [`security/examples/sample-pr.diff`](../examples/sample-pr.diff) → [`security/examples/sample-report.md`](../examples/sample-report.md) (pass 2). Index: [`security/examples/README.md`](../examples/README.md).
