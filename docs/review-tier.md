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
go run ./cmd/tier --incomplete "why"                      # change set unreadable: high
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
| `--incomplete REASON` | high, 2 humans, the reason is recorded; no diff is read (fail closed) |

Globs: `*` and `?` stay inside one path segment, `**` is a whole segment matching zero or more segments,
and a pattern without `/` matches the base name at any depth (`*.md`). `docs/**` is everything under
`docs/`. A `*.md` rule (low) loses to `security/skills/**` (medium) for `security/skills/x/SKILL.md`, because the highest
tier wins. The config rejects unknown fields and bad globs.

Config: [`.github/review-tiers.json`](../.github/review-tiers.json); adopters copy
`templates/review-tiers.json`.

## Enforcement (turn these on)

[`review-tier.yml`](../.github/workflows/review-tier.yml) labels the PR `risk:*` (and `needs-human`),
posts the tier, and runs a job named **`human-approvals`**. That job is an **advisory signal**: it shows
how many approvals the tier asks for and how many the **current head commit** has. It does not enforce
anything (see below). Only reviewers with admin, maintain or write permission count (checked per reviewer
with the collaborators permission API; if that call fails the reviewer is not counted and the log says so). Bot logins, `[bot]` accounts and the PR author never
count; a later "changes requested" from the same person cancels their approval. It re-runs on every review.
It always runs and fails closed: if `classify` was cancelled, failed or produced no tier, it fails, except
in the explicit bootstrap case below.

**Enforcement is native branch protection or a ruleset**, with "Require approvals", "Require review from
Code Owners" and "Dismiss stale approvals". Do not rely on `human-approvals` for this: on
`pull_request_review` GitHub runs the workflow file from the PR merge ref (forks too), so any PR can add its
own workflow with a job named `human-approvals`, and required checks cannot tell same-named checks apart.

Single maintainer? GitHub does not let the author approve their own PR, so a medium or high PR could
never pass. Set `"approvals": "advisory"` in the config and the job warns instead of failing; the tier,
labels and comment stay the same. The default (`"enforce"`, or no field) fails. This repo uses
`advisory` because it has one maintainer.

The workflow uses `pull_request_target` and reads the classifier and config from the **base** branch, so
a PR cannot lower its own tier. It never runs PR code and never fetches or checks out the PR head: the changed paths come from the
GitHub REST API (the PR files list) as data. If the API cannot give the full list (error, 3000 file cap,
count mismatch, control characters in a path) the PR is classified **high** with the reason recorded,
never low. When the base branch has no `cmd/tier` yet (the PR that introduces it), `classify` sets
`bootstrap=true`, emits a notice and skips classification, and `human-approvals` passes with a notice.
For `pull_request_target` events the workflow and classifier come from the base branch, so a PR cannot
change them for that event; the `pull_request_review` event is different (above). The classifier output is
checked against an allow-list (tier, humans, always, mode) before use; anything else fails the job.

Repository settings (Settings, Rules or Branches):

1. Require a pull request with the number of approvals the tier asks for (1 or 2), and require the
   status checks `go` and `review`. `human-approvals` may be listed too, but it is a signal, not the gate.
2. **Dismiss stale pull request approvals when new commits arrive.**
3. **Require review from Code Owners** ([`CODEOWNERS`](../.github/CODEOWNERS) lists the always-human and high paths).
4. Require a merge queue, so checks re-run against main with the change applied. Add the
   `merge_group` trigger to any workflow that must run there (`ci.yml` today only has `push` and `pull_request`).
5. Do not give the Actions bot or an agent account write access or approval rights. An agent that approves
   with a normal user account that has write access is not detected; keep such accounts out of the
   repository or add their logins to the jq filter in the workflow.

Limits: the classifier sees paths only; the second opinion (`reviewers/risk-second-opinion.md`) covers
what paths cannot. "The prompter of an agent-written PR counts as one reviewer" is a human rule
recorded in the PR template, not something the workflow can detect.
