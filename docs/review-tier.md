# Risk tier classifier

**English** · [Português (Brasil)](review-tier.pt-BR.md)

`cmd/tier` reads a diff and prints the risk tier and how many humans must approve. It is pure Go
(standard library), deterministic, and never calls a model. The change sets the tier, not the author:
[process/01-risk-tiers.md](../process/01-risk-tiers.md).

## Use

```bash
go run ./cmd/tier --from origin/main --to HEAD            # text
go run ./cmd/tier --from origin/main --to HEAD --format json --author someone
git diff main... | go run ./cmd/tier -                    # stdin
go run ./cmd/tier change.diff                             # file
```

Exit 0 classified, 2 usage, read, git or config error (a bad config never defaults to low). JSON:

```json
{"schemaVersion":1,"tier":"medium","humanApprovals":1,"alwaysHuman":false,"approvalMode":"enforce",
 "reasons":[{"path":"internal/x.go","rule":"medium: internal/**","tier":"medium"}]}
```

## Rules

| Rule | Effect |
|---|---|
| Highest path wins | Tier of the PR = riskiest changed path (renames and deletions count, both sides) |
| Tier to humans | low 0, medium 1, high 2 |
| No rule matches | `default` from the config (set it to `medium`) |
| `alwaysHuman` globs and `.github/workflows/**` | `alwaysHuman: true`, at least 1 human, whatever the tier |
| `--author` not in `trustedAuthors` | At least medium. Skipped when the list is empty or no author is given |
| Empty diff | low, 0 humans |

Globs: `*` and `?` stay inside one path segment, `**` is a whole segment matching zero or more segments,
and a pattern without `/` matches the base name at any depth (`*.md`). `docs/**` is everything under
`docs/`. A `*.md` rule (low) loses to `security/skills/**` (medium) for `security/skills/x/SKILL.md`, because the highest
tier wins. The config rejects unknown fields and bad globs.

Config: [`.github/review-tiers.json`](../.github/review-tiers.json); adopters copy
`templates/review-tiers.json`.

## Enforcement (turn these on)

[`review-tier.yml`](../.github/workflows/review-tier.yml) labels the PR `risk:*` (and `needs-human`),
posts the tier, and runs a job named **`human-approvals`** that fails until enough humans approved the
**current head commit**. Bot logins, `[bot]` accounts and the PR author never count; a later
"changes requested" from the same person cancels their approval. It re-runs on every review.

Single maintainer? GitHub does not let the author approve their own PR, so a medium or high PR could
never pass. Set `"approvals": "advisory"` in the config and the job warns instead of failing; the tier,
labels and comment stay the same. The default (`"enforce"`, or no field) fails. This repo uses
`advisory` because it has one maintainer.

The workflow uses `pull_request_target` and reads the classifier and config from the **base** branch, so
a PR cannot lower its own tier. It never runs PR code. Changes to the workflow itself take effect only
after merge.

Repository settings (Settings, Rules or Branches):

1. Require a pull request, and require the status checks `human-approvals`, `go`, `review`.
2. **Dismiss stale pull request approvals when new commits arrive.**
3. Require review from Code Owners ([`CODEOWNERS`](../.github/CODEOWNERS) lists the always-human and high paths).
4. Require a merge queue, so checks re-run against main with the change applied. Add the
   `merge_group` trigger to any workflow that must run there (`ci.yml` today only has `push` and `pull_request`).
5. Do not give the Actions bot or an agent account approval rights. An agent that approves with a
   normal user account is not detected; keep such accounts out of the repository or add their logins to
   the jq filter in the workflow.

Limits: the classifier sees paths only; the second opinion (`reviewers/risk-second-opinion.md`) covers
what paths cannot. "The prompter of an agent-written PR counts as one reviewer" is a human rule
recorded in the PR template, not something the workflow can detect.
