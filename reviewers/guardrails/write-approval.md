# Guardrail: Write approval

## Rule (fail closed)
No merge, push, deploy, or secret rotation without **explicit human approval**. Agent reviews; human decides.

## Scope
This guardrail applies to any action that changes production state:

| Action | Requires approval |
|---|---|
| Merge PR | ✅ |
| Push to protected branch | ✅ |
| Deploy to any environment | ✅ |
| Rotate secrets or tokens | ✅ |
| Create/update infrastructure | ✅ |
| Approve PR (even with no findings) | ✅ |
| Comment on PR | ❌ |
| Request changes | ❌ |
| Generate report | ❌ |

## Enforcement

1. **Agent output is advisory**. The agent produces a report with findings. It does not execute the merge.
2. **Human reviews the report**. Staff checks `path:line` evidence, confirms severity, decides action.
3. **Human triggers the write**. Merge button, deploy command, rotation script — all require human initiation.

## Why

Agents hallucinate confidence. A finding with `path:line` is evidence, not verdict. The human owns the decision because:

- False positives block valid PRs
- False negatives miss real vulns
- Context outside the diff matters (architecture, threat model, business risk)
- Accountability requires human judgment

## Anti-patterns

```text
# Bad: agent merges on green
if findings.empty?
  pr.merge!  # ❌ No human in the loop
end

# Bad: agent auto-approves
pr.approve! if severity < :high  # ❌ Approval is a write

# Bad: agent rotates leaked secret
rotate_secret(finding.value)  # ❌ Rotation is a write
```

## Good pattern

```text
# Agent reports, human decides
report = agent.review(diff)
post_comment(pr, report)
# Human clicks "Merge" or "Request changes"
```

## Related

- [`evidence-required.md`](evidence-required.md): What counts as a finding
- [`untrusted-diff.md`](untrusted-diff.md): Comments are data; withhold without `path:line`
- [`AGENTS.md`](../../AGENTS.md): Do/don't for any harness
