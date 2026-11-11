# Sample report — Path traversal (fixture)

Not a production incident. Walkthrough of [`security/skills/injection.md`](../skills/injection.md) against [`path-traversal.sample.diff`](path-traversal.sample.diff).

## Hunt
- Diff adds `download`: `req.params.name` is joined to `UPLOAD_DIR` and passed to `res.sendFile` with no base-directory check (hunt table: file path from request input).
- Diff adds `preview`: the same input is resolved, then checked with `startsWith(UPLOAD_DIR + path.sep)` before `res.sendFile` — anti-pattern table: **not** a finding.
- `logo` serves a fixed path → not a finding.

## Findings

```text
HIGH | src/api/files.ts:12 | req.params.name reaches res.sendFile through path.join with no normalization plus prefix check | CWE-22
```

## Evidence
- Source: `src/api/files.ts:11` — `file = path.join(UPLOAD_DIR, req.params.name)`. `path.join` keeps `..` segments that climb out of `UPLOAD_DIR`.
- Sink: `src/api/files.ts:12` — `res.sendFile(file)`.
- Explicit non-finding: `src/api/files.ts:16` and `src/api/files.ts:17` — `path.resolve` plus `startsWith` guard before the sink at `src/api/files.ts:20`.

## Verdict
**FAIL**

## Fix direction
Reuse the `preview` guard at `src/api/files.ts:12`: resolve the path and reject it unless it starts with `UPLOAD_DIR + path.sep`. Or call `res.sendFile(name, { root: UPLOAD_DIR })`, or keep only `path.basename(req.params.name)`.

## Out of scope / not claimed
- No claim about which files exist on the host or what the process can read.
- No CVSS, no exploit PoC, no payload.
- `health` is unchanged → not a finding.
- Agent does not patch the handler. Human decides request-changes.
