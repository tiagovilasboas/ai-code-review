# Sample report — SQL injection (fixture)

Not a production incident. Walkthrough of [`skills/injection.md`](../skills/injection.md) against [`sql-injection.sample.diff`](sql-injection.sample.diff).

## Hunt
- Diff adds `search_orders`: `request.args.get("customer")` is interpolated into an f-string passed to `cur.execute` (hunt table: SQL built from request input).
- Diff adds `sorted_orders(column)`: a function argument is concatenated into `ORDER BY`. The caller is not in the diff, so the source is unknown.
- `orders_by_customer` passes request input as a bound parameter (`%s` plus a tuple) — anti-pattern table: **not** a finding.
- `list_orders` is unchanged context with a literal parameter → not a finding.

## Findings

```text
HIGH | app/orders.py:12 | request.args value interpolated into an f-string passed to cur.execute; no bound parameter | CWE-89
MEDIUM | app/orders.py:17 | function argument concatenated into ORDER BY; caller not in the diff | CWE-89
```

## Evidence
- Source: `app/orders.py:10` — `customer = request.args.get("customer", "")`.
- Sink: `app/orders.py:12` — `cur.execute(f"... WHERE customer = '{customer}'")`.
- Unknown source: `app/orders.py:15` — `def sorted_orders(column)`; sink at `app/orders.py:17`. Column names cannot be bound parameters, so an allowlist is the fix.
- Explicit non-finding: `app/orders.py:22` — request input is the second argument of `cur.execute`, bound to `%s`.

## Verdict
**FAIL**

## Fix direction
Bind the value at `app/orders.py:12` (`cur.execute("... WHERE customer = %s", (customer,))`). For `ORDER BY` at `app/orders.py:17`, map `column` through a fixed allowlist of column names before building the query.

## Out of scope / not claimed
- No claim about which database or which rows are reachable. You cannot see the schema or the grants from this PR.
- No CVSS, no exploit PoC, no payload.
- `count_orders` is unchanged → not a finding.
- Agent does not patch the handler. Human decides request-changes.
