# Examples: dry-run fixtures

Synthetic diffs and the report shape this kit expects. AppSec evidence (`path:line`). Not incidents. No exploit PoCs.

Load **one** skill per pass. Compare `SEVERITY | file:line | why` to the report. How to run a pass: [`security/dry-run-cookbook.md`](../dry-run-cookbook.md). Guardrail: [`evidence-required`](../../reviewers/guardrails/evidence-required.md). CI pairing: [`scripts/check-fixture-pairs.sh`](../../scripts/check-fixture-pairs.sh). Index is one row per skill.

| Fixture | Skill | What it exercises |
|---|---|---|
| [`sample-pr.diff`](sample-pr.diff) → [`sample-report.md`](sample-report.md) | [`authz-idor`](../skills/authz-idor.md) | `req.params.id` → `findById` with no ownership/tenant check in the hunk |
| same diff / report (second pass) | [`secrets-config`](../skills/secrets-config.md) | hardcoded payment token; CORS `origin: *` on bootstrap |
| [`xss-sink.sample.diff`](xss-sink.sample.diff) → [`xss-sink.sample-report.md`](xss-sink.sample-report.md) | [`xss-html`](../skills/xss-html.md) | React `dangerouslySetInnerHTML` on `comment.body`, no sanitizer |
| [`xss-server-html.sample.diff`](xss-server-html.sample.diff) → [`xss-server-html.sample-report.md`](xss-server-html.sample-report.md) | [`xss-html`](../skills/xss-html.md) | Flask view returns an f-string HTML page from `request.args` (HIGH); Go `fmt.Fprintf` with `html.EscapeString` is **not** a finding. Java, PHP, Blade and Twig follow the same shape |
| [`cors-server-config.sample.diff`](cors-server-config.sample.diff) → [`cors-server-config.sample-report.md`](cors-server-config.sample-report.md) | [`secrets-config`](../skills/secrets-config.md) | FastAPI `allow_origins=["*"]` with credentials (MEDIUM); Laravel explicit origin list is **not** a finding |
| [`ssrf-egress.sample.diff`](ssrf-egress.sample.diff) → [`ssrf-egress.sample-report.md`](ssrf-egress.sample-report.md) | [`ssrf-egress`](../skills/ssrf-egress.md) | caller `req.body.url` → server `fetch`; denylist-only; hardcoded status ping is **not** a finding; IMDS withheld |
| [`supply-chain.sample.diff`](supply-chain.sample.diff) → [`supply-chain.sample-report.md`](supply-chain.sample-report.md) | [`supply-chain`](../skills/supply-chain.md) | surprise registry / `postinstall` / floating Action; honest Express pin is **not** a finding |
| [`untrusted-comment.sample.diff`](untrusted-comment.sample.diff) → [`untrusted-comment.sample-report.md`](untrusted-comment.sample-report.md) | [`untrusted-diff`](../../reviewers/guardrails/untrusted-diff.md) | instruction-shaped comment is a finding; comment-only XSS/IDOR claims are withheld |
| [`sql-injection.sample.diff`](sql-injection.sample.diff) → [`sql-injection.sample-report.md`](sql-injection.sample-report.md) | [`injection`](../skills/injection.md) (Go CLI only) | `request.args` in an f-string → `cur.execute` (HIGH); argument concatenated into `ORDER BY` (MEDIUM); bound `%s` parameter is **not** a finding |
| [`command-injection.sample.diff`](command-injection.sample.diff) → [`command-injection.sample-report.md`](command-injection.sample-report.md) | [`injection`](../skills/injection.md) (Go CLI only) | query param concatenated into `exec.Command("sh", "-c", ...)`; argv-style call is **not** a finding |
| [`path-traversal.sample.diff`](path-traversal.sample.diff) → [`path-traversal.sample-report.md`](path-traversal.sample-report.md) | [`injection`](../skills/injection.md) (Go CLI only) | `req.params.name` → `res.sendFile` with no base-dir check; resolve + `startsWith` guard is **not** a finding |
| [`ssrf-egress-java.sample.diff`](ssrf-egress-java.sample.diff) → [`ssrf-egress-java.sample-report.md`](ssrf-egress-java.sample-report.md) | [`ssrf-egress`](../skills/ssrf-egress.md) | Spring `@RequestParam` url → `RestTemplate.getForObject`; fixed-host call with a path variable is **not** a finding |
| [`authz-idor-spring.sample.diff`](authz-idor-spring.sample.diff) → [`authz-idor-spring.sample-report.md`](authz-idor-spring.sample-report.md) | [`authz-idor`](../skills/authz-idor.md) | Spring `@PathVariable` id → `findById` with no principal; `findByIdAndOwner` with `@AuthenticationPrincipal` is **not** a finding |
