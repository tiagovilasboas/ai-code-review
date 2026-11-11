#!/usr/bin/env bash
# review-noise.sh <owner/repo> [--since YYYY-MM-DD] [--bot <login>] [--max-prs N]
#
# Read-only. Classifies the review threads a bot opened on recent PRs, and
# reports rounds and findings per PR. Needs gh (authenticated) and jq. No LLM.
# See process/03-noise-metrics.md for how to read the numbers.
#
# A "finding" is a review thread whose first comment is by the bot. --bot LOGIN
# restricts to that login (e.g. claude, github-actions[bot]); the default is any
# author GitHub types as Bot.
#
# Classes (HEURISTIC, from the thread state and keywords in human replies; the
# first rule that matches wins, so expect misclassification and sample it):
#   wrong        a human reply says false positive / not a bug / incorrect / wrong
#   intentional  a human reply says intentional / by design / on purpose / expected
#   didnt-matter a human reply says nit / won't fix / out of scope / follow-up,
#                or the thread was resolved with no human reply and not outdated
#   fixed        the thread is outdated (the commented lines changed after the
#                comment) or a reply says fixed / done / addressed / good catch
#   unanswered   open, not outdated, no human reply
#   other        a human replied with none of the keywords, thread still open
# "Outdated" also happens when a line is rewritten for an unrelated reason, so
# fixed is an upper bound. Keywords are English only.
#
# rounds: highest "Round N of 3" found in bot reviews/comments on the PR; if the
# bot never wrote one, the number of bot reviews. findings: bot-opened threads.
# Limits: first 100 threads, 50 reviews and 50 comments per PR; PR list capped by
# --max-prs (default 100, newest updated first).
set -euo pipefail

usage() { echo "usage: $0 <owner/repo> [--since YYYY-MM-DD] [--bot <login>] [--max-prs N]" >&2; exit 2; }
[ $# -ge 1 ] || usage
repo=$1; shift
case $repo in */*) ;; *) usage ;; esac
since="1970-01-01"; bot=""; max=100
while [ $# -gt 0 ]; do
  case $1 in
    --since) [ $# -ge 2 ] || usage; since=$2; shift 2 ;;
    --bot) [ $# -ge 2 ] || usage; bot=$2; shift 2 ;;
    --max-prs) [ $# -ge 2 ] || usage; max=$2; shift 2 ;;
    *) usage ;;
  esac
done
case $since in [0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]) ;; *) usage ;; esac
case $max in '' | *[!0-9]*) usage ;; esac
command -v gh >/dev/null && command -v jq >/dev/null || { echo "gh and jq are required" >&2; exit 2; }
owner=${repo%%/*}; name=${repo#*/}

query='query($owner:String!,$name:String!,$cursor:String){
 repository(owner:$owner,name:$name){pullRequests(first:25,after:$cursor,orderBy:{field:UPDATED_AT,direction:DESC}){
  pageInfo{hasNextPage endCursor}
  nodes{number updatedAt
   reviews(first:50){nodes{body author{login __typename}}}
   comments(first:50){nodes{body author{login __typename}}}
   reviewThreads(first:100){nodes{isResolved isOutdated
    comments(first:20){nodes{body author{login __typename}}}}}}}}}'

tmp=$(mktemp); trap 'rm -f "$tmp"' EXIT
cursor=""; count=0
while :; do
  args=(-f query="$query" -F owner="$owner" -F name="$name")
  [ -n "$cursor" ] && args+=(-F cursor="$cursor")
  page=$(gh api graphql "${args[@]}")
  echo "$page" | jq -c --arg since "$since" '.data.repository.pullRequests.nodes[] | select(.updatedAt[0:10] >= $since)' >>"$tmp"
  count=$((count + $(echo "$page" | jq '.data.repository.pullRequests.nodes | length')))
  oldest=$(echo "$page" | jq -r '.data.repository.pullRequests.nodes[-1].updatedAt[0:10] // "0000-00-00"')
  more=$(echo "$page" | jq -r '.data.repository.pullRequests.pageInfo.hasNextPage')
  cursor=$(echo "$page" | jq -r '.data.repository.pullRequests.pageInfo.endCursor')
  [ "$more" = true ] && [ "$count" -lt "$max" ] && [ "$oldest" \> "$since" -o "$oldest" = "$since" ] || break
done

jq -s --arg bot "$bot" --arg repo "$repo" --arg since "$since" '
def isbot: (.author.__typename == "Bot") or ($bot != "" and .author.login == $bot);
def mine: if $bot != "" then (.author.login == $bot) else (.author.__typename == "Bot") end;
def hit($t; $re): $t | test($re; "i");
def classify:
  ([.comments.nodes[1:][] | select(isbot | not) | .body] | join("\n")) as $h
  | if   hit($h; "false positive|not a bug|incorrect|\\bwrong\\b|misread|does(n.t| not) apply") then "wrong"
    elif hit($h; "intentional|by design|on purpose|deliberate|expected behavio") then "intentional"
    elif hit($h; "\\bnit\\b|won.t fix|wontfix|not worth|out of scope|follow-up|doesn.t matter|irrelevant") then "didnt-matter"
    elif .isOutdated or hit($h; "\\bfixed\\b|\\bdone\\b|addressed|good catch") then "fixed"
    elif $h != "" then (if .isResolved then "fixed" else "other" end)
    elif .isResolved then "didnt-matter"
    else "unanswered" end;
[ .[] | . as $pr
  | [ $pr.reviewThreads.nodes[] | select(.comments.nodes[0] | mine) ] as $ft
  | ([ ($pr.reviews.nodes + $pr.comments.nodes)[] | select(isbot) | .body
       | scan("Round ([0-9]+) of 3") | .[0] | tonumber ]) as $r
  | ([ $pr.reviews.nodes[] | select(isbot) ] | length) as $nrev
  | select(($ft | length) > 0 or $nrev > 0)
  | { pr: $pr.number, findings: ($ft | length),
      rounds: (if ($r | length) > 0 then ($r | max) else $nrev end),
      classes: ([ $ft[] | classify ] | group_by(.) | map({key: .[0], value: length}) | from_entries) } ] as $prs
| { repo: $repo, since: $since, bot: (if $bot == "" then "any Bot" else $bot end),
    prs: $prs,
    totals: ( ["fixed","wrong","intentional","didnt-matter","unanswered","other"] as $ks
      | [ $prs[].classes ] as $cs
      | { prsWithBotReview: ($prs | length),
          findings: ([ $prs[].findings ] | add // 0),
          meanFindingsPerPr: (if ($prs|length) > 0 then ((([ $prs[].findings ] | add // 0) / ($prs|length)) * 100 | round / 100) else 0 end),
          meanRoundsPerPr: (if ($prs|length) > 0 then ((([ $prs[].rounds ] | add // 0) / ($prs|length)) * 100 | round / 100) else 0 end),
          prsWith5PlusRounds: ([ $prs[] | select(.rounds >= 5) ] | length),
          classes: ( reduce $ks[] as $k ({}; .[$k] = ([ $cs[][$k] // 0 ] | add // 0)) ) } ) }
' "$tmp" >"$tmp.json"

tab=$(printf '\t')
jq -r '"Repo \(.repo)  since \(.since)  bot: \(.bot)\n"' "$tmp.json"
jq -r '"PR\tfindings\trounds\tclasses", (.prs[] | "#\(.pr)\t\(.findings)\t\(.rounds)\t\(.classes | to_entries | map("\(.key)=\(.value)") | join(" "))")' "$tmp.json" | column -t -s "$tab"
echo
jq -r '"class\tcount", (.totals.classes | to_entries[] | "\(.key)\t\(.value)")' "$tmp.json" | column -t -s "$tab"
jq -r '"\nPRs with bot review: \(.totals.prsWithBotReview)  findings: \(.totals.findings)  mean findings/PR: \(.totals.meanFindingsPerPr)  mean rounds/PR: \(.totals.meanRoundsPerPr)  PRs with 5+ rounds: \(.totals.prsWith5PlusRounds)\n(heuristic classes; see the script header)\n"' "$tmp.json"
jq . "$tmp.json"
rm -f "$tmp.json"
