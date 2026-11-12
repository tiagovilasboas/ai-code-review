# Risk tiers and human approval

[Português (Brasil)](01-risk-tiers.md) · **English**

**The change sets the tier, not the author.** A deterministic classifier assigns the tier from the paths touched; a second opinion asks whether this change needs a person regardless of tier.

| Tier | Typical changes | Human approvals |
|---|---|---|
| low | Docs, copy, internal-only code, additive behaviour behind a flag | 0, once AI review and security are green |
| medium | Ordinary product changes; sensitive areas such as caching, auto-updater, system prompts | 1 |
| high | Highest-risk code (billing, auth, migrations) plus a small set of paths that always need two | 2 |

## Rules
| Rule | Effect |
|---|---|
| Highest path wins | One high path makes the whole PR high |
| Second opinion | [`reviewers/risk-second-opinion.md`](../reviewers/risk-second-opinion.md) can only **raise** the need for a person, never lower a tier |
| `alwaysHuman` globs | Owner-marked sensitive paths always need a person, even if tier is low |
| CI workflow files | `.github/workflows/**` always needs a person |
| Trusted authors | The light path (low, 0 humans) is only for logins in `trustedAuthors`. Untrusted author means at least medium, 1 human |
| Bot approval never counts | Agent and bot approvals are not human approvals |
| Prompter counts | The person who prompted an agent-written PR is one required reviewer, not a bystander |
| Commit-specific | Approvals are dismissed on new commits; re-approve the final head |

## Enforcement in this kit
- **Classify:** `go run ./cmd/tier --from origin/main --to HEAD --config .github/review-tiers.json` prints
  `{"schemaVersion":1,"tier":"medium","humanApprovals":1,"alwaysHuman":false,"approvalMode":"enforce","reasons":[{"path":"...","rule":"...","tier":"..."}]}`. Exit 0 ok, 2 usage/read error. Every tier has `reasons` with paths, so a tier is never unexplained.
- **Report:** [`.github/workflows/review-tier.yml`](../.github/workflows/review-tier.yml) runs `cmd/tier` and posts the tier and reasons on the PR.
- **Make it binding:** native branch protection or a ruleset with "Require approvals" (1 or 2), "Require review from Code Owners" (CODEOWNERS for owner-marked paths) and "Dismiss stale approvals". The tier tool tells you the number; branch protection enforces it. The workflow's `human-approvals` job is an advisory signal only: on `pull_request_review` GitHub runs the workflow from the PR merge ref, so a PR can ship a same-named job.
- **Config:** [`.github/review-tiers.json`](../.github/review-tiers.json) here; adopters start from [`templates/review-tiers.json`](../templates/review-tiers.json).

## Zero-cost fallback
`cmd/tier` is pure Go, no model. Without branch-protection rulesets (private repos on some plans), treat the tier comment as a required manual checklist item.

## Failure modes
| Failure | Guard |
|---|---|
| Tier too low | Second opinion, `alwaysHuman`, default `medium` for unmatched paths |
| Author trust abused | Trust is a per-repo list, reviewed like code |
| Rubber-stamp approvals | Humans read evidence ([05](05-human-roles.en.md)), not the agent summary |

Next: [02 convergence](02-convergence.en.md).
