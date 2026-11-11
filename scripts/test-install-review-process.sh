#!/bin/sh
# Exercises scripts/install-review-process.sh: dry-run, install, idempotency, refusal to overwrite, security kit.
set -eu
HERE=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
INSTALL="$HERE/install-review-process.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
FAILS=0

ok() { printf 'ok   %s\n' "$1"; }
fail() { printf 'FAIL %s\n' "$1"; FAILS=$((FAILS + 1)); }
check() { # check <name> <command...>
  name=$1; shift
  if "$@" >/dev/null 2>&1; then ok "$name"; else fail "$name"; fi
}

# 1. dry-run writes nothing
D="$TMP/dry"
sh "$INSTALL" --harness all --target "$D" --dry-run >"$TMP/dry.out"
check "dry-run prints a plan" grep -q '^create    reviewers/bug-hunter.md' "$TMP/dry.out"
check "dry-run writes no files" test -z "$(find "$D" -type f 2>/dev/null | head -1)"

# 2. real install, every harness
T="$TMP/repo"
sh "$INSTALL" --harness all --target "$T" >"$TMP/i1.out"
for f in reviewers/bug-hunter.md reviewers/verifier.md reviewers/security.md reviewers/risk-classifier.md \
  reviewers/risk-second-opinion.md reviewers/pre-pr-self-check.md reviewers/convergence.md \
  reviewers/guardrails/evidence-required.md reviewers/guardrails/untrusted-diff.md \
  reviewers/guardrails/write-approval.md \
  .agents/skills/pr-review-process/SKILL.md \
  .claude/agents/bug-hunter.md .claude/commands/review-pr.md .claude/commands/pre-pr.md \
  .kiro/steering/review-process.md .kiro/hooks/review-pr.kiro.hook \
  .codex/prompts/review-pr.md AGENTS.md \
  .cursor/rules/review-process.mdc \
  .github/copilot-instructions.md .github/prompts/review-pr.prompt.md \
  GEMINI.md .gemini/commands/review-pr.toml; do
  check "installed $f" test -f "$T/$f"
done
check "README of adapters not installed" test ! -e "$T/README.md"
check "hooks.example.json not installed" test ! -e "$T/hooks.example.json"
check "security kit not installed without the flag" test ! -e "$T/security"
check "no guardrails/ at the target root" test ! -e "$T/guardrails"

# 3. idempotent: second run changes nothing and exits 0
sum1=$(cd "$T" && find . -type f | sort | xargs cksum)
sh "$INSTALL" --harness all --target "$T" >"$TMP/i2.out"
sum2=$(cd "$T" && find . -type f | sort | xargs cksum)
check "second run exits 0 and is stable" test "$sum1" = "$sum2"
check "snippet block appended once" test "$(grep -c 'review-process:begin' "$T/AGENTS.md")" = 1

# 4. refuses to overwrite a modified file, keeps it, exits non-zero
echo "my edits" >"$T/reviewers/bug-hunter.md"
rc=0
sh "$INSTALL" --harness claude-code --target "$T" >"$TMP/i3.out" || rc=$?
check "conflict exits 3" test "$rc" = 3
check "modified file kept" grep -q 'my edits' "$T/reviewers/bug-hunter.md"
check "conflict reported" grep -q '^conflict  reviewers/bug-hunter.md' "$TMP/i3.out"

# 5. --force overwrites
sh "$INSTALL" --harness claude-code --target "$T" --force >/dev/null
check "--force restores file" grep -q 'Bug hunter' "$T/reviewers/bug-hunter.md"

# 6. existing AGENTS.md is appended to, never clobbered
U="$TMP/existing"
mkdir -p "$U"
echo "# Mine" >"$U/AGENTS.md"
sh "$INSTALL" --harness codex --target "$U" >/dev/null
check "existing AGENTS.md preserved" grep -q '^# Mine' "$U/AGENTS.md"
check "snippet appended to existing AGENTS.md" grep -q 'review-process:begin' "$U/AGENTS.md"

# 7. --with-security-kit adds the security skills and the appsec skill
S="$TMP/kit"
sh "$INSTALL" --harness claude-code --target "$S" --with-security-kit >/dev/null
for f in security/skills/authz-idor.md security/skills/cloudflare-config.md \
  .agents/skills/appsec-pr-review/SKILL.md reviewers/guardrails/untrusted-diff.md; do
  check "security kit installed $f" test -f "$S/$f"
done
check "no skills/ at the target root" test ! -e "$S/skills"

# 7b. --force with a begin marker and no end marker refuses and keeps the file byte for byte
M="$TMP/noend"
mkdir -p "$M"
printf '# Mine\n<!-- review-process:begin -->\npartial\ntail kept?\n' >"$M/AGENTS.md"
cp "$M/AGENTS.md" "$TMP/noend.orig"
rc=0
sh "$INSTALL" --harness codex --target "$M" --force >"$TMP/noend.out" 2>&1 || rc=$?
check "begin without end under --force exits 3" test "$rc" = 3
check "begin without end leaves file unchanged" cmp -s "$M/AGENTS.md" "$TMP/noend.orig"
check "begin without end names the missing marker" grep -q 'review-process:end' "$TMP/noend.out"

# 8. bad args
rc=0; sh "$INSTALL" --harness nope --target "$TMP/x" >/dev/null 2>&1 || rc=$?
check "unknown harness exits 2" test "$rc" = 2

if [ "$FAILS" -gt 0 ]; then printf '%s check(s) failed\n' "$FAILS"; exit 1; fi
echo "all install checks passed"
