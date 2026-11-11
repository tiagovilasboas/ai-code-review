# CLI contract (v1)

```
review [--format text|json|sarif] [--output FILE] <diff-file>
review [--format ...] [--output FILE] --git [--from REF --to REF | --commit SHA] [--dir DIR]
```

- Default format: `text` (pinned byte for byte by `internal/cli/testdata/golden`).
- `--output FILE` writes the report to FILE; stdout gets nothing on success.
- `--git` with no range: workspace changes (staged, unstaged, untracked) vs HEAD.
- `--from A --to B`: changes of B since merge-base with A. `--commit SHA`: one commit.
- Exit codes: 0 no BLOCK, 1 any BLOCK, 2 usage, read or git error.
- Diff sources live in `internal/source`; output formats in `internal/report`.

## JSON (schemaVersion 1)

```json
{"schemaVersion":1,"tool":{"name":"ai-code-review","version":"<semver|dev>"},
 "decision":{"merge":false,"reason":"AuthZ/IDOR at src/api/orders.ts:36 (ASVS-5.0-8.2.2)","blocker":"src/api/orders.ts:36"},
 "findings":[{"title":"...","class":"AuthZ/IDOR","findingClass":"authz","path":"...","line":36,
   "cwe":"CWE-639","owasp":["A01:2021"],"severity":"HIGH","action":"BLOCK"}]}
```

## SARIF 2.1.0

- `tool.driver.name` = `ai-code-review`; one rule per finding class; `ruleId` = finding class.
- `level`: BLOCK → `error`, REVIEW → `warning`.
- Rule `properties`: `tags` (`security`, CWE, OWASP IDs), `security-severity`
  (CRITICAL 9.5, HIGH 8.0, MEDIUM 5.5, LOW 3.0) so GitHub code scanning ranks it.
- Result location: `artifactLocation.uri` = path, `region.startLine` = line.
- `partialFingerprints.primaryLocationLineHash` = stable hash of class+path+line text.
