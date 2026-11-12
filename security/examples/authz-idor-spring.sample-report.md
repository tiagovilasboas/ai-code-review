# Sample report — AuthZ / IDOR, Spring (fixture)

Not a production incident. Walkthrough of [`security/skills/authz-idor.md`](../skills/authz-idor.md) against [`authz-idor-spring.sample.diff`](authz-idor-spring.sample.diff).

## Findings

```text
HIGH | src/main/java/app/OrderController.java:7 | @PathVariable id passed to findById with no principal, ownership filter or @PreAuthorize in the handler | CWE-639
```

## Evidence
- Identifier: `src/main/java/app/OrderController.java:6` — `@PathVariable Long id`.
- Sink: `src/main/java/app/OrderController.java:7` — `orders.findById(id)`.
- Explicit non-finding: `src/main/java/app/OrderController.java:12` — `findByIdAndOwner(id, me.getUsername())` binds the lookup to the authenticated principal.
- Withhold: ownership enforced in unseen filters or a class-level annotation would be **insufficient evidence**; none is in the hunk.

## Verdict
**FAIL** — AuthZ/IDOR HIGH, request changes. Human decides; the agent does not patch or merge.

## Out of scope / not claimed
- No CVSS, no exploit PoC, no claim about unseen middleware.
