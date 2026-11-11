# AI review workflow (opt-in)

**English** · [Português (Brasil)](ai-review-workflow.pt-BR.md)

[`ai-review.yml`](../.github/workflows/ai-review.yml) runs the multi-agent review of
[`pr-review-process`](../.agents/skills/pr-review-process/SKILL.md) with `anthropics/claude-code-action`.
It is advice and blocks nothing. It does not run the deterministic engine; for a blocking gate, run the root
[`action.yml`](../action.yml) with `fail-on-block: true` as a required check.

## Turn it on

1. Add the repository secret `ANTHROPIC_API_KEY`. Without it the job prints a skip message and passes.
2. Keep `REVIEW.md` current. It is the reviewer's contract (depth, nit cap, skip paths, evidence). The
   reviewer reads it from the base branch, so a change takes effect once merged.
3. Optionally run it by hand: Actions, ai-review, Run workflow, with a PR number. A manual run posts the
   summary only: the pinned action installs its inline-comment tool for pull request events alone.

It never runs on fork PRs or drafts. Permissions are `contents: read`, plus `pull-requests: write` and
`issues: write` for comments and labels.

## Rounds

| Round | What the reviewer reads |
|---|---|
| 1 | The full diff against the whole repo and the PR's intent |
| 2, 3 | Only the commits since the last reviewed SHA; a posted finding is never repeated |
| 4+ | No review. The workflow asks the author to reconsider the approach |

State lives on the PR: label `ai-review:round-N` and a hidden marker comment holding the last reviewed
SHA. Delete the labels to reset the count. Reports end with `Round N of 3`, which
[`scripts/review-noise.sh`](../scripts/review-noise.sh) reads.

## Security

The diff under review is written by the PR author, so anything in it may try to steer the reviewer.
The workflow treats the whole PR as data and keeps secrets and instructions out of its reach.

- **No shell.** Trusted steps run before Claude and write its inputs to files: the diff
  (`git diff --no-ext-diff --no-textconv`), and the PR title, body and commit messages. Claude gets only
  `Read`, `Grep`, `Glob`, subagents and the action's inline-comment tool; `Bash`, `Edit`, `Write` and the
  web tools are denied. Only numbers, validated SHAs and runner paths enter the prompt.
- **Reads stay in the checkout.** `--restricted` confines the file tools to the checkout and the inputs
  folder, and deny rules also block `/proc`, the runner temp folder and every `.git/` (the action writes its
  token into `.git/config`). `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` turns on the CLI's best-effort secret scrub
  for subprocesses.
- **Trusted output.** Claude returns the report as structured output. A later step reads it through `env:`,
  redacts token-shaped strings and posts it. Claude has no tool that writes PR comments other than inline ones.
- **Base config and instructions.** Before Claude starts, `.claude/`, `.mcp.json`, `.claude.json`,
  `CLAUDE.md`, `CLAUDE.local.md`, `.gitmodules`, `.ripgreprc` and `.husky` are reset to the base commit
  (deleted if the base has none), and the CLI runs with `--setting-sources user --strict-mcp-config`, so a PR
  cannot add hooks or MCP servers. The reviewer follows `SKILL.md`, `REVIEW.md`, `AGENTS.md` and
  `reviewers/` from a worktree of the base commit, so a PR cannot rewrite its own review.
- **Round state.** The last reviewed SHA comes only from a comment whose whole body matches this workflow's
  marker, and it is used only if it is still an ancestor of the head; otherwise the round reviews the whole PR.
- **What is left.** Injected text can still bias findings or post a misleading inline comment, and an encoded
  secret would pass the redaction if a read ever got through. The review is advice; a person decides.
- The action is pinned to a full SHA (v1.0.271); its inputs, outputs and flags were checked against that
  release. Check them again before bumping.

## Cost and noise

Run `scripts/review-noise.sh owner/repo --since YYYY-MM-DD` monthly
([process/03](../process/03-noise-metrics.md)). Its classes are heuristics, so sample them.

## Nightly incident evals

[`review-evals.yml`](../.github/workflows/review-evals.yml) runs `go run ./cmd/bench` every night and
fails if true positives, misses, false positives, silence on clean cases or any class score regress
against `bench/results.json`. Add a corpus case for every incident, then refresh the baseline with
`make bench` in the same PR ([process/04](../process/04-incident-evals.md)). It scores the deterministic
engine, not the LLM reviewer.
