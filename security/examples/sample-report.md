# Sample report — Stage 1 dry-run

Harness-agnostic example of running this kit on [`sample-pr.diff`](sample-pr.diff).
Every finding has `file:line`. No exploit PoC. Agent does not merge.

## Pass 1 — `security/skills/authz-idor.md`

```text
HIGH | src/api/orders.ts:36 | orderId from req.params.id passed to findById without ownership/tenant check in the diff
```

Notes for the human:

- Route registration in `src/server.ts` adds `GET /orders/:id` but does not show auth middleware in the hunk → do not invent middleware; the sink at `orders.ts:36` is enough for the AuthZ skill.
- If ownership were enforced only in unseen middleware, a second review with a larger diff might downgrade to `insufficient evidence`.

## Pass 2 — `security/skills/secrets-config.md`

```text
CRITICAL | config/app.ts:15 | hardcoded payment API token literal in source
MEDIUM | src/server.ts:8 | CORS origin set to * on app bootstrap
```

Notes for the human:

- Redact secret values in any paste to chat/tickets; the sample token is a placeholder (`pk_test_REDACTED_SAMPLE_ONLY_0000`).
- Agent must not rotate keys or edit the PR. Human decides: request changes, scrub history, rotate if a real secret ever appeared.

## Aggregate (copy-paste style)

```text
HIGH | src/api/orders.ts:36 | orderId from path used without ownership check
CRITICAL | config/app.ts:15 | hardcoded payment API token in source
MEDIUM | src/server.ts:8 | CORS origin set to *
```

## HITL decision

| Option | When |
|---|---|
| Request changes | Any CRITICAL/HIGH with clear `file:line` in the PR under review |
| Merge with follow-up | LOW only, or insufficient evidence documented |
| Escalate | Real secret in a shared remote — rotate outside this kit |

Agent stops here. Human merges or not.
