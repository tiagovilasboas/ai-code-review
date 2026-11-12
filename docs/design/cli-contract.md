# CLI contract (v1)

```
review [--format text|json|sarif] [--output FILE] <diff-file>
review [--format ...] [--output FILE] --git [--from REF --to REF | --commit SHA] [--dir DIR]
review ... [--triage FILE]   # either form: apply external triage verdicts
```

- Default format: `text` (pinned byte for byte by `internal/cli/testdata/golden`).
- `--output FILE` writes the report to FILE; stdout gets nothing on success.
- `--git` with no range: workspace changes (staged, unstaged, untracked) vs HEAD.
- `--from A --to B`: changes of B since merge-base with A. `--commit SHA`: one commit.
- Exit codes: 0 no BLOCK, 1 any BLOCK, 2 usage, read or git error.
- `--triage FILE` applies a verdicts file produced outside the binary (see Triage). Any problem with the file exits 2 with no report.
- Diff sources live in `internal/source`; output formats in `internal/report`.

## JSON (schemaVersion 1)

```json
{"schemaVersion":1,"tool":{"name":"ai-code-review","version":"<semver|dev>"},
 "decision":{"merge":false,"reason":"AuthZ/IDOR at src/api/orders.ts:36 (ASVS-5.0-8.2.2)","blocker":"src/api/orders.ts:36"},
 "findings":[{"title":"...","class":"AuthZ/IDOR","findingClass":"authz","path":"...","line":36,
   "cwe":"CWE-639","owasp":["A01:2021"],"severity":"HIGH","action":"BLOCK",
   "id":"3f9a1c0d5b7e2a64"}]}
```

`id` is the stable finding id (first 8 bytes of sha256 of `class|path|line|cwe`, hex). It is the key triage verdicts use.
With `--triage`, the report gains a `triage` object (absent otherwise), after `findings`:

```json
"triage":{"applied":[{"id":"3f9a1c0d5b7e2a64","verdict":"downgrade","reason":"...","class":"XSS / HTML sink",
  "findingClass":"xss","path":"a.ts","line":3,"fromSeverity":"HIGH","fromAction":"BLOCK",
  "toSeverity":"MEDIUM","toAction":"REVIEW"}]}
```

`toSeverity`/`toAction` are omitted for `drop`. `decision` and `findings` reflect the result after triage.

## SARIF 2.1.0

- `tool.driver.name` = `ai-code-review`; one rule per finding class; `ruleId` = finding class.
- `level`: BLOCK → `error`, REVIEW → `warning`.
- Rule `properties`: `tags` (`security`, CWE, OWASP IDs), `security-severity`
  (CRITICAL 9.5, HIGH 8.0, MEDIUM 5.5, LOW 3.0) so GitHub code scanning ranks it.
- Result location: `artifactLocation.uri` = path, `region.startLine` = line.
- `partialFingerprints.primaryLocationLineHash` = stable hash of class+path+line text.

## Triage (verdicts schema 1)

The engine never calls a model. An optional LLM pass runs outside the binary
(procedure: `reviewers/triage.md`) and writes a verdicts file; `--triage FILE` only applies it.

```json
{"schemaVersion":1,"verdicts":[
  {"id":"3f9a1c0d5b7e2a64","action":"drop","reason":"test placeholder at a.ts:3"},
  {"id":"9c2e77a1b04d58f3","action":"downgrade","severity":"MEDIUM","reason":"constant input at b.ts:40"}]}
```

- Strict decoding: unknown fields, trailing data, a `schemaVersion` other than 1, a missing/oversized (500 bytes) `reason`, duplicate ids, `drop` with a `severity`, `downgrade` without a valid one.
- `action` is `drop` or `downgrade`; nothing else exists, so a verdict can never add a finding. An id not in the report is rejected.
- `downgrade` must strictly lower severity (CRITICAL > HIGH > MEDIUM > LOW). The action follows the usual mapping (CRITICAL/HIGH BLOCK, else REVIEW), so HIGH to MEDIUM turns BLOCK into REVIEW and can flip the exit code to 0.
- Immune classes: `untrusted-diff` (CWE-1427, the prompt-injection signal itself) and `secret` (always CRITICAL; the model sees the value, and a placeholder is a rule fix). A verdict naming one is an error.
- Fail closed: an unreadable file, bad JSON, a rejected verdict, an ambiguous id (two findings share it) or two verdicts for one id exits 2, prints `error: triage: ...` to stderr and no report. Never a partial apply.
- Visibility: text appends a `Triaged:` section (verdict, finding, ids, before/after, reason); JSON adds `triage.applied`; SARIF keeps each dropped finding as a result with `suppressions` (`kind: external`, `status: accepted`, `justification` = reason) and puts `properties.triage` on dropped and downgraded results. A run without `--triage` is byte-identical to before in text and SARIF; JSON only gained `findings[].id`.
- Exit code is computed after triage.
