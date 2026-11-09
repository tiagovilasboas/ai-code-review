#!/bin/sh
# Runs the ai-code-review CLI for the composite action and writes step outputs.
# All inputs arrive via environment variables; nothing is interpolated by the workflow.
#   ACTION_PATH       action checkout (module root)         [required]
#   FORMAT            sarif|json|text                        [default sarif]
#   WANT_SARIF        true|false: also produce SARIF         [default false]
#   BASE_REF          base branch/sha (empty: BASE_REF_DEFAULT, then HEAD^)
#   BASE_REF_DEFAULT  PR base branch from the event
#   DIFF_FILE         review this diff file instead of git mode (self-test)
#   WORKSPACE         repository to review                   [default $GITHUB_WORKSPACE]
#   OUT_DIR           report directory                       [default $RUNNER_TEMP/ai-code-review]
# Exit: 0 reviewed (see outputs.blocked), 2 usage/read/git/validation error.
set -eu

FORMAT=${FORMAT:-sarif}
WANT_SARIF=${WANT_SARIF:-false}
WORKSPACE=${WORKSPACE:-${GITHUB_WORKSPACE:-.}}
OUT_DIR=${OUT_DIR:-${RUNNER_TEMP:-/tmp}/ai-code-review}
OUTPUT_FILE=${GITHUB_OUTPUT:-/dev/null}
SUMMARY_FILE=${GITHUB_STEP_SUMMARY:-/dev/null}
DIFF_FILE=${DIFF_FILE:-}
: "${ACTION_PATH:?ACTION_PATH is required}"

die() { echo "::error::$*" >&2; exit 2; }

case $FORMAT in sarif | json | text) ;; *) die "format must be sarif, json or text" ;; esac
mkdir -p "$OUT_DIR"

bin="$OUT_DIR/review"
(cd "$ACTION_PATH" && go build -o "$bin" ./cmd/review) || die "build failed"

wgit() { git -C "$WORKSPACE" "$@"; }
is_shallow() { [ "$(wgit rev-parse --is-shallow-repository)" = true ]; }

from=""
if [ -z "$DIFF_FILE" ]; then
  base=${BASE_REF:-${BASE_REF_DEFAULT:-}}
  if [ -n "$base" ]; then
    # Safe ref charset, no leading dash (option injection), no "..".
    case $base in -* | *..* | *[!A-Za-z0-9._/-]*) die "base-ref contains unsafe characters" ;; esac
    if is_shallow; then
      wgit fetch --no-tags --unshallow origin >&2 || die "cannot deepen shallow clone"
    fi
    wgit fetch --no-tags origin "+refs/heads/$base:refs/remotes/origin/$base" >&2 || true
    if wgit rev-parse --verify --quiet "refs/remotes/origin/$base^{commit}" >/dev/null; then
      from="origin/$base"
    elif wgit rev-parse --verify --quiet "$base^{commit}" >/dev/null; then
      from="$base"
    else
      die "cannot resolve base-ref"
    fi
  else
    if is_shallow; then
      wgit fetch --no-tags --deepen=1 origin >&2 || die "cannot deepen shallow clone"
    fi
    from=$(wgit rev-parse --verify "HEAD^") || die "no base-ref and no parent commit"
  fi
fi

# run_format FMT FILE: sets rc; exit code 2 (or worse) aborts, 1 means BLOCK.
run_format() {
  rc=0
  if [ -n "$DIFF_FILE" ]; then
    "$bin" --format "$1" --output "$2" "$DIFF_FILE" || rc=$?
  else
    "$bin" --format "$1" --output "$2" --git --from "$from" --to HEAD --dir "$WORKSPACE" || rc=$?
  fi
  [ "$rc" -le 1 ] || die "review failed with exit code $rc"
}

json="$OUT_DIR/report.json"
run_format json "$json"
blocked=false
[ "$rc" -eq 1 ] && blocked=true

sarif_file=""
if [ "$FORMAT" = sarif ] || [ "$WANT_SARIF" = true ]; then
  sarif_file="$OUT_DIR/report.sarif"
  run_format sarif "$sarif_file"
fi
if [ "$FORMAT" = text ]; then
  run_format text "$OUT_DIR/report.txt"
fi

findings=$(jq '.findings | length' "$json")

{
  echo "## ai-code-review"
  echo
  if [ "$blocked" = true ]; then
    echo "**DO NOT MERGE**: $(jq -r '.decision.reason' "$json")"
  else
    echo "No BLOCK findings. Findings: $findings"
  fi
  if [ "$findings" -gt 0 ]; then
    echo
    echo "| Action | Severity | Class | Location | Title |"
    echo "| --- | --- | --- | --- | --- |"
    jq -r '.findings[] | "| \(.action) | \(.severity) | \(.class) | `\(.path):\(.line)` | \(.title | gsub("\\|"; "\\|") | gsub("[\\r\\n]+"; " ")) |"' "$json"
  fi
} >>"$SUMMARY_FILE"

{
  echo "blocked=$blocked"
  echo "findings=$findings"
  echo "sarif-file=$sarif_file"
} >>"$OUTPUT_FILE"
