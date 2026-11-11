#!/usr/bin/env bash
# Pair security/examples/*.diff with expected reports and require path:line on findings.
# Naming:
#   name.sample.diff  <->  name.sample-report.md
#   sample-pr.diff    <->  sample-report.md   (Stage 1 names)
#   other.diff        <->  other-report.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLES="${ROOT}/security/examples"
SEVERITY='CRITICAL|HIGH|MEDIUM|LOW'
# path:line (GitHub alert style). Allows .npmrc and nested paths.
PATH_LINE_RE='[A-Za-z0-9_.@+-]+(/[A-Za-z0-9_.@+-]+)*:[0-9]+'
FINDING_LINE_RE="^[[:space:]]*[-*]?[[:space:]]*\`?(${SEVERITY})\`?[[:space:]]*\\|"

failures=0
paired_reports=()

fail() {
  printf 'FAIL | %s\n' "$1" >&2
  failures=$((failures + 1))
}

ok() {
  printf 'OK   | %s\n' "$1"
}

report_for_diff() {
  local base
  base="$(basename "$1")"
  case "${base}" in
    sample-pr.diff) printf '%s\n' "sample-report.md" ;;
    *.sample.diff) printf '%s\n' "${base%.sample.diff}.sample-report.md" ;;
    *) printf '%s\n' "${base%.diff}-report.md" ;;
  esac
}

# Print "start end" for each new-file hunk of path in a unified diff.
hunk_ranges_for_path() {
  local diff_file="$1"
  local want="$2"
  local in_file=0
  local line start count end

  while IFS= read -r line || [[ -n "${line}" ]]; do
    if [[ "${line}" == "+++ b/${want}" ]]; then
      in_file=1
      continue
    fi
    if [[ "${line}" == diff\ --git* ]]; then
      in_file=0
      continue
    fi
    if [[ "${line}" == +++\ * ]]; then
      in_file=0
      continue
    fi
    if (( in_file )) && [[ "${line}" =~ ^@@\ -[0-9]+(,[0-9]+)?\ \+([0-9]+)(,([0-9]+))?\ @@ ]]; then
      start="${BASH_REMATCH[2]}"
      count="${BASH_REMATCH[4]:-1}"
      if [[ "${start}" == "0" || "${count}" == "0" ]]; then
        continue
      fi
      end=$((start + count - 1))
      printf '%s %s\n' "${start}" "${end}"
    fi
  done < "${diff_file}"
}

line_in_hunks() {
  local lineno="$1"
  local start end
  while read -r start end; do
    if (( lineno >= start && lineno <= end )); then
      return 0
    fi
  done
  return 1
}

strip_ticks() {
  printf '%s\n' "${1//\`/}"
}

if [[ ! -d "${EXAMPLES}" ]]; then
  fail "security/examples/: directory missing"
  exit 1
fi

shopt -s nullglob
diffs=("${EXAMPLES}"/*.diff)
if (( ${#diffs[@]} == 0 )); then
  fail "security/examples/: no .diff fixtures"
  exit 1
fi

for diff_path in "${diffs[@]}"; do
  report_name="$(report_for_diff "${diff_path}")"
  report_path="${EXAMPLES}/${report_name}"
  rel_diff="security/examples/$(basename "${diff_path}")"
  rel_report="security/examples/${report_name}"
  paired_reports+=("${report_name}")

  if [[ ! -f "${report_path}" ]]; then
    fail "${rel_diff}: missing paired report ${rel_report}"
    continue
  fi

  finding_lines=()
  while IFS= read -r line || [[ -n "${line}" ]]; do
    if [[ "${line}" =~ ${FINDING_LINE_RE} ]]; then
      finding_lines+=("${line}")
    fi
  done < "${report_path}"

  if (( ${#finding_lines[@]} == 0 )); then
    if grep -Eiq 'insufficient evidence|no evidence-based findings' "${report_path}"; then
      ok "${rel_diff} -> ${rel_report} (insufficient evidence / no findings)"
      continue
    fi
    fail "${rel_report}: no SEVERITY | path:line finding and no insufficient-evidence marker"
    continue
  fi

  pair_ok=1
  for raw in "${finding_lines[@]}"; do
    plain="$(strip_ticks "${raw}")"
    if [[ ! "${plain}" =~ ${PATH_LINE_RE} ]]; then
      fail "${rel_report}: finding lacks path:line: ${plain}"
      pair_ok=0
      continue
    fi

    cite="${BASH_REMATCH[0]}"
    cite_path="${cite%:*}"
    cite_line="${cite##*:}"

    if [[ ! -s "${diff_path}" ]]; then
      fail "${rel_diff}: empty diff (cannot locate ${cite})"
      pair_ok=0
      continue
    fi

    ranges="$(hunk_ranges_for_path "${diff_path}" "${cite_path}")"
    if [[ -z "${ranges}" ]]; then
      fail "${rel_report}: ${cite} path not in a hunk of ${rel_diff}"
      pair_ok=0
      continue
    fi
    if ! printf '%s\n' "${ranges}" | line_in_hunks "${cite_line}"; then
      fail "${rel_report}: ${cite} line is outside hunks of ${rel_diff}"
      pair_ok=0
    fi
  done

  if (( pair_ok == 1 )); then
    ok "${rel_diff} -> ${rel_report} (${#finding_lines[@]} finding(s) with path:line in hunk)"
  fi
done

for report_path in "${EXAMPLES}"/*report.md; do
  report_name="$(basename "${report_path}")"
  orphan=1
  for paired in "${paired_reports[@]}"; do
    if [[ "${paired}" == "${report_name}" ]]; then
      orphan=0
      break
    fi
  done
  if (( orphan == 1 )); then
    fail "security/examples/${report_name}: report is not paired to any security/examples/*.diff"
  fi
done

if (( failures > 0 )); then
  printf '\n%d pairing/evidence check(s) failed\n' "${failures}" >&2
  exit 1
fi

printf '\nAll fixture pairs have path:line evidence where required.\n'
