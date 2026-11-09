# GitHub Action

Composite action that builds the Go CLI from the pinned action checkout, reviews the
pull request diff (`--git --from <base> --to HEAD`), writes a findings table to the job
summary and uploads SARIF to Code scanning. It follows the CLI contract in
[design/cli-contract.md](design/cli-contract.md).

## Usage

```yaml
name: ai-code-review
on: pull_request

permissions:
  contents: read
  security-events: write # SARIF upload; drop it with upload-sarif: "false"

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          fetch-depth: 0
      - uses: tiagovilasboas/ai-code-review@<full-commit-sha>
```

Pin the action to a full commit SHA, like any third-party action.

## Inputs

| Input | Default | Meaning |
| --- | --- | --- |
| `format` | `sarif` | Report format of the output file: `sarif`, `json` or `text`. |
| `upload-sarif` | `true` | Upload SARIF to Code scanning (category `ai-code-review`). |
| `fail-on-block` | `true` | Fail the job when any finding is BLOCK. |
| `base-ref` | PR base branch | Base branch or commit; falls back to `HEAD^` outside pull requests. |
| `go-version-file` | action's `go.mod` | `go.mod` read by `actions/setup-go`. |
| `comment-summary` | `true` | Write the findings table to the job summary. |

## Outputs

| Output | Meaning |
| --- | --- |
| `blocked` | `true` when at least one finding is BLOCK. |
| `findings` | Number of findings. |
| `sarif-file` | Path of the SARIF report; empty when none was produced. |

## Permissions

- `contents: read` always.
- `security-events: write` only when `upload-sarif` is `true`.
- No `pull-requests: write`: the action never comments on the PR. Results go to the
  job summary and Code scanning.

## Exit behavior

- Exit 1 from the CLI (BLOCK) sets `blocked=true`. SARIF is still uploaded, then the
  job fails if `fail-on-block` is `true`.
- Exit 2 (usage, read, git or build error, unsafe `base-ref`) always fails the step.

## Code scanning

Findings appear under Security > Code scanning with the tool name `ai-code-review`,
one rule per finding class, `error` for BLOCK and `warning` for REVIEW, ranked by
`security-severity` (CRITICAL 9.5, HIGH 8.0, MEDIUM 5.5, LOW 3.0). On pull requests they
are annotated on the changed lines.

## Security notes

- Inputs and event fields reach the scripts only through `env:` and are quoted.
- `base-ref` must match `[A-Za-z0-9._/-]+`, not start with `-` and not contain `..`.
- Only the Go toolchain and module proxy are used; the module has no runtime dependencies.

## Limitations

- Needs full history for merge-base: use `fetch-depth: 0`. A shallow clone is
  deepened automatically, which costs a full fetch.
- Pull requests from forks get a read-only token, so SARIF upload fails there; set
  `upload-sarif: "false"` for fork PRs and rely on the summary and exit code.
- Code scanning on private repositories requires GitHub Advanced Security.
- Runners need `jq` (present on `ubuntu-latest`) and a POSIX `sh`.
- The review is deterministic and diff-only: it does not analyze unchanged code.
