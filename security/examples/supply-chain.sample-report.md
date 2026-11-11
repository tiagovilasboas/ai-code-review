# Sample report — supply-chain install path (fixture)

Not a production incident. Walkthrough of [`security/skills/supply-chain.md`](../skills/supply-chain.md) against [`supply-chain.sample.diff`](supply-chain.sample.diff).

## Hunt
- `package.json` adds an unexplained `postinstall` (hunt table: install-time script).
- New `.npmrc` points `registry` at `pkgs.example.invalid` (not the usual npmjs host).
- Workflow on `pull_request` adds third-party `some-org/cache-warmup@v1` (floating tag) and replaces `npm ci` with `npm install`.
- `package-lock.json` only pins `express` `4.19.1` → `4.19.2` on `registry.npmjs.org` with an integrity field — anti-pattern table: **not** a finding.

## Findings

```text
HIGH | package.json:8 | new postinstall runs node scripts/fetch-tool.js at npm ci; script body not in the hunk
HIGH | .npmrc:2 | registry host changed to pkgs.example.invalid
HIGH | .github/workflows/ci.yml:9 | third-party Action some-org/cache-warmup@v1 (moving tag) on pull_request
MEDIUM | .github/workflows/ci.yml:14 | CI uses npm install, which can rewrite the lockfile
```

## Evidence
- Install script: `package.json:8` — `postinstall`. `scripts/fetch-tool.js` is not in the diff → do not invent what it downloads; the script entry is enough to FAIL.
- Registry: `.npmrc:2` — `registry=https://pkgs.example.invalid/npm/`.
- Floating Action: `.github/workflows/ci.yml:9` — `uses: some-org/cache-warmup@v1`. Trigger `pull_request` is in the same file at `.github/workflows/ci.yml:3`. First-party `actions/checkout@v4` at line 8 is not a finding.
- Lockfile rewrite: `.github/workflows/ci.yml:14` — `run: npm install` replaces `npm ci`.
- Explicit non-finding: `package-lock.json:13–15` — `resolved` stays on `registry.npmjs.org`, integrity present. Do not invent a CVE for Express.

## Verdict
**FAIL** (install path, not SCA)

## Fix direction
Drop or explain `postinstall` at `package.json:8`. Point `.npmrc` at the registry you meant (npmjs or a known mirror). Pin `cache-warmup` to a commit SHA, or drop the third-party Action. CI: `npm ci`, not `npm install`. The Express lockfile pin can stay.

## Out of scope / not claimed
- No CVE, CVSS, or “this package is malware”.
- No named incident (event-stream, xz, …) claimed here.
- `scripts/fetch-tool.js` body is outside the diff → not a second finding.
- First-party `actions/checkout@v4` / `actions/setup-node@v4` accepted at a tag for this smoke.
- Agent does not merge, pin, or rewrite the lockfile. Human decides request-changes.
