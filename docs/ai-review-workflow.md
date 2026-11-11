# AI review workflow (opt-in)

**English** · [Português (Brasil)](ai-review-workflow.pt-BR.md)

[`ai-review.yml`](../.github/workflows/ai-review.yml) runs the multi-agent review of
[`pr-review-process`](../.agents/skills/pr-review-process/SKILL.md) with `anthropics/claude-code-action`.
It is advice on top of the deterministic gate; the blocking security check remains the `review` job in `ci.yml`.

## Turn it on

1. Add the repository secret `ANTHROPIC_API_KEY`. Without it the job prints a skip message and passes.
2. Keep `REVIEW.md` current. It is the reviewer's contract (depth, nit cap, skip paths, evidence).
3. Optionally run it by hand: Actions, ai-review, Run workflow, with a PR number.

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

## Safety

- The PR title, body and diff are untrusted data. Only numbers and validated SHAs enter the prompt.
- The tool list is read-only plus inline comments; the agent cannot merge, approve or edit.
- The action is pinned to a full SHA (v1.0.271). Check its inputs before bumping.
- Cost and noise are yours to watch: run `scripts/review-noise.sh owner/repo --since YYYY-MM-DD`
  monthly ([process/03](../process/03-noise-metrics.md)). Its classes are heuristics, so sample them.

## Nightly incident evals

[`review-evals.yml`](../.github/workflows/review-evals.yml) runs `go run ./cmd/bench` every night and
fails if true positives, misses, false positives, silence on clean cases or any class score regress
against `bench/results.json`. Add a corpus case for every incident, then refresh the baseline with
`make bench` in the same PR ([process/04](../process/04-incident-evals.md)). It scores the deterministic
engine, not the LLM reviewer.
