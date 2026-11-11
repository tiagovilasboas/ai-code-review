# Skill — Supply chain (PR smoke)

## Purpose
Catch install-time and lockfile surprises in a PR: new registries, install scripts, unpinned Actions, `curl | sh`. You are not running SCA. You are reading the diff.

## Risk class
Cite these IDs only. Do not invent CWE / CVE / CVSS / extra catalog numbers.

| Kind | IDs (official) | Why this skill |
|---|---|---|
| OWASP Top 10:2021 | [A06:2021](https://owasp.org/Top10/2021/A06_2021-Vulnerable_and_Outdated_Components/) Vulnerable and Outdated Components; [A08:2021](https://owasp.org/Top10/2021/A08_2021-Software_and_Data_Integrity_Failures/) Software and Data Integrity Failures | Lockfile honesty vs install-path integrity (scripts, registry, unsigned Actions) |
| OWASP ASVS 5.0 V15 | [15.1.2](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x24-V15-Secure-Coding-and-Architecture.md) components from pre-defined, trusted, maintained repositories; [15.2.4](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x24-V15-Secure-Coding-and-Architecture.md) expected repository / no dependency confusion | Registry host, lockfile `resolved`, install scripts |
| Agentic posture | [ASI04:2026](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/) Agentic Supply Chain Vulnerabilities; [AST02](https://owasp.org/www-project-agentic-skills-top-10/) Supply Chain Compromise | If the PR also adds agent skills / MCP, say so. Do not invent a CVE from memory. |

CLI prints `A08:2021, ASVS-5.0-15.2.4` for install-path integrity and `A06:2021, ASVS-5.0-15.1.2` for lockfile-rewrite CI. ASI/AST stay here.

## Hunt table

| Signal in the diff | What you are looking for | Evidence to cite |
|---|---|---|
| `package.json` / `requirements.txt` / `go.mod` / `Gemfile` / `Cargo.toml` | New dep, version pin removed, range widened to `*` / `latest` | manifest `path:line` |
| `package-lock.json` / `yarn.lock` / `pnpm-lock.yaml` / `poetry.lock` | `resolved` host ≠ expected registry; integrity hash dropped; lockfile rewritten “for no app reason” | lockfile `path:line` |
| `.npmrc` / `.yarnrc*` / pip `index-url` / `PIP_EXTRA_INDEX_URL` | Registry or extra-index pointed at a new host | config `path:line` |
| `postinstall` / `preinstall` / `install` scripts in `package.json` | Arbitrary shell at `npm ci` | script `path:line` |
| CI: `npm install` instead of `npm ci` (or yarn/pnpm without frozen lockfile) | Lockfile can be rewritten on CI | workflow `path:line` |
| GitHub Action `uses: org/action@v1` (moving tag) vs `@<sha>` | Third-party CI code you did not pin | workflow `path:line` |
| `curl … \| sh` / `wget … \| bash` in Dockerfile or setup | Unsigned remote installer | Dockerfile/`*.sh` `path:line` |
| New `vendor/`, git submodule URL, or `http://` tarball | Code you did not review, fetched over cleartext | fetch site `path:line` |

A lockfile that only bumps a known package to a specific version, with `resolved` on the usual registry and an integrity hash, is usually **not** a finding. Say that.

## Anti-patterns (look-alikes)

Hunt is the signal. This table is the trap: do not promote a look-alike to a CVE, and do not drop a real install-path FAIL.

| In the hunk | Usually **not** a finding | Usually **FAIL** (cite `path:line`) |
|---|---|---|
| Lockfile only; `resolved` is npmjs (or your known mirror); `integrity` present | Version pin / bump. Optional note: “lockfile only, expected registry.” | `resolved` host changed, or `integrity` removed |
| `package.json` uses `^` / `~` and the lockfile is committed | Manifest range is normal; the lockfile is the pin | Range widened to `*` / `latest` **and** lockfile dropped or CI runs bare `npm install` |
| `uses: actions/checkout@v4` (first-party GitHub) | Accept the tag if that is house policy — say so | `uses: third-party/action@v1` (moving tag) on `pull_request` or with secrets |
| `postinstall` that only runs a file **in this repo** and the script is in the hunk | LOW / explain at `path:line` if it is a local compile | New `postinstall` / `preinstall` you cannot explain; `curl \| sh` / unsigned remote installer |
| `.npmrc` `registry=https://registry.npmjs.org/` | Expected host | New host, extra-index, or `always-auth` aimed at a surprise registry |
| “This package had a CVE last year” | — | **Never** a finding by itself. No version in the diff → `insufficient evidence`. Do not invent CWE-1104. |

## Required evidence
Every finding: **`path:line`** of the manifest, lockfile field, workflow step, or installer line.

Do not paste a CVE from memory. If the diff does not show the package/version, you do not have a vuln — you have a hygiene note, or insufficient evidence.

Guardrail: [`reviewers/guardrails/evidence-required.md`](../../reviewers/guardrails/evidence-required.md).

## Pass / fail
- **FAIL** — new/changed install script; registry/index host changed; lockfile `resolved` to an unexpected host or missing integrity; CI install that rewrites the lockfile; Action pinned to a floating tag *and* it runs on `pull_request` from forks or has secrets; `curl | sh`.
- **PASS** — deps change, lockfile stays on the expected registry with integrity, CI uses `npm ci` / frozen lockfile, Actions pinned to commit SHA (or first-party `actions/*` you are willing to trust at a tag — say which).
- **INSUFFICIENT** — “this package had a CVE last year” with no version in the diff. Do not invent CWE-1104.

This skill does **not** prove the new version is safe. It proves the *install path* in the PR is honest.

## Fix direction
1. Pin what you install. Commit the lockfile. CI: `npm ci` / `--frozen-lockfile`, not a bare `install` that rewrites it. See [npm ci](https://docs.npmjs.com/cli/v10/commands/npm-ci).
2. Keep `resolved` on the registry you meant (npmjs, your mirror). Treat extra-index / surprise `.npmrc` as a review event.
3. Drop `postinstall` unless you can explain the script at `path:line`.
4. Pin GitHub Actions to a commit SHA. Tags move. [GitHub: security hardening for Actions](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions#using-third-party-actions).
5. No pipe-to-shell. Vendor or use a signed package.

## Out of scope
- Full SBOM, SCA, or “are we on a vulnerable lodash”. That is a scanner job; this is a PR skill.
- Claiming a named incident (xz, event-stream, etc.) happened *here*.
- Provenance/SLSA attestations unless the workflow in the diff actually verifies them — then cite the step.
- Malware analysis of a tarball.

## Output format
- `SEVERITY` | `file:line` | summary | (optional CWE)
- Name the host/script/action. Do not dump the whole lockfile.

## Refs (public)
- [OWASP Top 10:2021 A06 Vulnerable and Outdated Components](https://owasp.org/Top10/2021/A06_2021-Vulnerable_and_Outdated_Components/)
- [OWASP Top 10:2021 A08 Software and Data Integrity Failures](https://owasp.org/Top10/2021/A08_2021-Software_and_Data_Integrity_Failures/)
- [OWASP ASVS 5.0 V15 Secure Coding and Architecture](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x24-V15-Secure-Coding-and-Architecture.md) (`15.1.2`, `15.2.4`)
- [CWE-1104](https://cwe.mitre.org/data/definitions/1104.html) / [CWE-829](https://cwe.mitre.org/data/definitions/829.html) — only when the install line is in the hunk

Worked fixture: [`security/examples/supply-chain.sample.diff`](../examples/supply-chain.sample.diff) → [`security/examples/supply-chain.sample-report.md`](../examples/supply-chain.sample-report.md). Index: [`security/examples/README.md`](../examples/README.md).
