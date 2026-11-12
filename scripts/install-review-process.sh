#!/bin/sh
# Install the multi-agent review process into a target repository.
# Usage: install-review-process.sh --harness <name|all> --target <dir> [--dry-run] [--force] [--with-security-kit]
# Harnesses: claude-code kiro codex cursor copilot gemini all
# Never overwrites an existing file unless --force. --dry-run prints the plan only.
set -eu

SELF_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
SRC=$(CDPATH='' cd -- "$SELF_DIR/.." && pwd)
HARNESS=""
TARGET=""
DRY=0
FORCE=0
SECKIT=0
CONFLICTS=0
BEGIN='<!-- review-process:begin -->'
END='<!-- review-process:end -->'

usage() {
  echo "Usage: $0 --harness <claude-code|kiro|codex|cursor|copilot|gemini|all> --target <dir> [--dry-run] [--force] [--with-security-kit]" >&2
  exit 2
}

while [ $# -gt 0 ]; do
  case $1 in
    --harness) [ $# -ge 2 ] || usage; HARNESS=$2; shift 2 ;;
    --target) [ $# -ge 2 ] || usage; TARGET=$2; shift 2 ;;
    --dry-run) DRY=1; shift ;;
    --force) FORCE=1; shift ;;
    --with-security-kit) SECKIT=1; shift ;;
    -h|--help) usage ;;
    *) echo "unknown argument: $1" >&2; usage ;;
  esac
done
[ -n "$HARNESS" ] && [ -n "$TARGET" ] || usage

ALL="claude-code kiro codex cursor copilot gemini"
if [ "$HARNESS" = all ]; then
  LIST=$ALL
else
  case " $ALL " in
    *" $HARNESS "*) LIST=$HARNESS ;;
    *) echo "unknown harness: $HARNESS" >&2; usage ;;
  esac
fi

if [ "$DRY" -eq 0 ]; then mkdir -p "$TARGET"; fi
if [ -d "$TARGET" ]; then TARGET=$(CDPATH='' cd -- "$TARGET" && pwd); fi

say() { printf '%s\n' "$*"; }

# copy_file <src> <rel-dst>
copy_file() {
  src=$1; dst="$TARGET/$2"
  if [ -e "$dst" ]; then
    if cmp -s "$src" "$dst"; then say "skip      $2 (identical)"; return 0; fi
    if [ "$FORCE" -eq 0 ]; then
      say "conflict  $2 (exists; use --force to overwrite)"
      CONFLICTS=$((CONFLICTS + 1)); return 0
    fi
    say "overwrite $2"
  else
    say "create    $2"
  fi
  [ "$DRY" -eq 1 ] && return 0
  mkdir -p "$(dirname -- "$dst")"
  cp "$src" "$dst"
}

# snippet_file <src> <rel-dst>: append a marked block; never clobber the file.
snippet_file() {
  src=$1; dst="$TARGET/$2"
  if [ ! -e "$dst" ]; then
    say "create    $2 (snippet)"
    [ "$DRY" -eq 1 ] && return 0
    mkdir -p "$(dirname -- "$dst")"; cp "$src" "$dst"; return 0
  fi
  if grep -qF "$BEGIN" "$dst"; then
    if [ "$FORCE" -eq 0 ]; then say "skip      $2 (block already present)"; return 0; fi
    if ! grep -qF "$END" "$dst"; then
      say "conflict  $2 (begin marker without end marker; file left untouched)"
      echo "error: $2 has '$BEGIN' but no '$END'; add the end marker or remove the block, then re-run" >&2
      return 0
    fi
    say "replace   $2 (snippet block)"
    [ "$DRY" -eq 1 ] && return 0
    tmp="$dst.review-process.tmp"
    awk -v b="$BEGIN" -v e="$END" -v f="$src" '
      index($0,b){skip=1; while((getline l < f)>0) print l; next}
      index($0,e){skip=0; next}
      !skip{print}' "$dst" > "$tmp" && mv "$tmp" "$dst"
    return 0
  fi
  say "append    $2 (snippet block)"
  [ "$DRY" -eq 1 ] && return 0
  { printf '\n'; cat "$src"; } >> "$dst"
}

# copy_tree <src-dir> <rel-dst-dir>: every file below, relative paths kept.
copy_tree() {
  base=$1; out=$2
  [ -d "$base" ] || { say "warn      missing source dir $base"; return 0; }
  (cd "$base" && find . -type f | sort) | while IFS= read -r f; do
    rel=${f#./}
    copy_file "$base/$rel" "$out/$rel"
  done
}

run_plan() {
  copy_tree "$SRC/reviewers" "reviewers"
  copy_file "$SRC/.agents/skills/pr-review-process/SKILL.md" ".agents/skills/pr-review-process/SKILL.md"
  copy_file "$SRC/.agents/skills/ai-code-review-setup/SKILL.md" ".agents/skills/ai-code-review-setup/SKILL.md"
  if [ -f "$SRC/templates/REVIEW.md" ]; then
    copy_file "$SRC/templates/REVIEW.md" "REVIEW.md"
  else
    say "warn      templates/REVIEW.md not found; REVIEW.md not installed"
  fi
  if [ -f "$SRC/templates/review-tiers.json" ]; then
    copy_file "$SRC/templates/review-tiers.json" ".github/review-tiers.json"
  else
    say "warn      templates/review-tiers.json not found; tiers config not installed"
  fi
  if [ "$SECKIT" -eq 1 ]; then
    copy_tree "$SRC/security/skills" "security/skills"
    copy_file "$SRC/.agents/skills/appsec-pr-review/SKILL.md" ".agents/skills/appsec-pr-review/SKILL.md"
  fi
  for h in $LIST; do
    hdir="$SRC/harness/$h"
    [ -d "$hdir" ] || { say "warn      missing harness dir harness/$h"; continue; }
    # files inside dot directories are installed as-is
    (cd "$hdir" && find . -type f | sort) | while IFS= read -r f; do
      rel=${f#./}
      case $rel in
        .*/*) copy_file "$hdir/$rel" "$rel" ;;
      esac
    done
    # snippets: <NAME>.snippet.md -> destination from the table
    case $h in
      codex) snippet_file "$hdir/AGENTS.snippet.md" "AGENTS.md" ;;
      copilot) snippet_file "$hdir/copilot-instructions.snippet.md" ".github/copilot-instructions.md" ;;
      gemini) snippet_file "$hdir/GEMINI.snippet.md" "GEMINI.md" ;;
    esac
  done
}

[ "$DRY" -eq 1 ] && say "dry-run: nothing will be written (target: $TARGET)"
OUT=$(run_plan)
printf '%s\n' "$OUT"
CONFLICTS=$(printf '%s\n' "$OUT" | grep -c '^conflict ' || true)
if [ "$CONFLICTS" -gt 0 ]; then
  if [ "$FORCE" -eq 1 ]; then
    say "$CONFLICTS file(s) kept as they were. Fix the markers listed above, then re-run."
  else
    say "$CONFLICTS file(s) kept as they were. Re-run with --force to overwrite."
  fi
  exit 3
fi
if [ "$DRY" -eq 0 ]; then
  say "done. Next: review REVIEW.md, edit .github/review-tiers.json, then run /review-pr (or the equivalent for your harness)."
fi
