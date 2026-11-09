# Benchmark

A reproducible precision/recall measurement of the deterministic engine
(`review.Run` with `review.Default`) on a labelled corpus of PR diffs.

```
make bench        # or: go run ./cmd/bench [--corpus bench/corpus] [--json bench/results.json]
```

The command prints a markdown table and writes `bench/results.json`. It exits 0
whatever the scores (it is a measurement, not a gate) and fails only if the
corpus is malformed or the results file cannot be written. Output is
deterministic: no clock, no network.

## Method

- Each case is a directory `bench/corpus/<case-id>/` with `change.diff` and
  `expected.json` (see [bench/README.md](../bench/README.md)).
- Expected findings are what a careful AppSec reviewer would flag in that diff.
  They are **not** derived from the engine's output, and cases the engine gets
  wrong are kept on purpose.
- Cases pair true positives with hard true negatives: sanitized or
  parameterized code, ownership-checked lookups, pinned actions, comment-only
  mentions, literal values and test fixtures. An empty `findings` list means the
  engine must stay silent.

## Matching rule

A reported finding matches an expected one only if **path, line and class** are
all equal (line is the new-file line). Duplicates on either side count once.

- TP: expected and reported.
- FN: expected, not reported (counted against the expected class).
- FP: reported, not expected (counted against the reported class).
- Precision = TP/(TP+FP), recall = TP/(TP+FN), F1 = harmonic mean. `n/a` means
  the denominator is zero. There are no true-negative counts; "clean cases
  silent" reports how many cases with no expected finding produced no output.

## Add a case

1. Create `bench/corpus/<class>-NN-short-name/`.
2. Write `change.diff` as a unified diff (new-file hunks are enough; line
   numbers come from the `@@ -a,b +c,d @@` headers).
3. Write `expected.json` with the lines a reviewer would flag, using new-file
   line numbers. Do not look at the engine's output while labelling.
4. Run `make bench`, commit the case and the refreshed `bench/results.json`,
   and update the table below.

Write cases yourself; do not copy code from other projects.

## Results

Three numbers, kept apart on purpose. Only the blind set estimates accuracy on
code the rules have not seen.

### 1. Baseline (before any tuning)

Run on 2026-11-09 at commit `20c617c`, on the first 111 cases, before any rule
was changed against this corpus.

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 3 | 2 | 2 | 0.60 | 0.60 | 0.60 |
| cors | 1 | 1 | 3 | 0.50 | 0.25 | 0.33 |
| install-script | 4 | 1 | 1 | 0.80 | 0.80 | 0.80 |
| lockfile-rewrite | 1 | 1 | 3 | 0.50 | 0.25 | 0.33 |
| registry-host | 3 | 0 | 4 | 1.00 | 0.43 | 0.60 |
| secret | 4 | 2 | 3 | 0.67 | 0.57 | 0.62 |
| ssrf | 4 | 3 | 2 | 0.57 | 0.67 | 0.62 |
| unpinned-action | 5 | 2 | 0 | 0.71 | 1.00 | 0.83 |
| untrusted-diff | 4 | 2 | 3 | 0.67 | 0.57 | 0.62 |
| xss | 5 | 5 | 2 | 0.50 | 0.71 | 0.59 |
| **overall** | 34 | 19 | 23 | 0.64 | 0.60 | 0.62 |

111 cases (57 expected findings, 54 clean cases of which 35 stayed silent).

### 2. Tuned corpus (`bench/corpus`, fitted)

The precision work then changed the rules while looking at these cases, and
added held-out `*-h*` cases written before each fix. The rules now pass every
case, which shows the fixes are general enough for this corpus and nothing more:
these numbers are **fitted, not an accuracy estimate**. Path-traversal, SQL and
command injection have few or no cases here; they are measured in the blind set.

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cors | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| install-script | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| lockfile-rewrite | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| registry-host | 13 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| secret | 12 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| ssrf | 10 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| unpinned-action | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| untrusted-diff | 13 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| xss | 12 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| **overall** | 101 | 0 | 0 | 1.00 | 1.00 | 1.00 |

### 3. Blind set (pending)

Not run yet. A separate author, with no access to the rules, the corpus or the
examples, is writing a held-out set under `bench/blind/`. Its results will be
added here when it lands. Until then, there is no accuracy estimate for code the
rules have not seen.

### Notable misses (baseline)

False negatives: Fastify `request.params` and Next app-router `params.id`
lookups (authz); manual `Access-Control-Allow-Origin: *` headers, reflected
origins and `next.config.js` headers (cors); `npm i`, `yarn install` and
multi-line `run:` blocks (lockfile-rewrite); look-alike registry hosts, yarn.lock
and pnpm-lock resolution, tarball dependencies (registry-host); connection
strings with passwords, PEM keys, literal npm tokens (secret); `axios({url})`
and `http.get` (ssrf); paraphrased or non-comment injection text
(untrusted-diff); jQuery `.html()` and `javascript:` hrefs (xss).

False positives: `DOMPurify.sanitize(...)` inside `__html` (the sanitizer
regex does not allow the `.sanitize` member), sanitized-earlier variables,
constant or empty `innerHTML` and test fixtures (xss); allowlisted, env-based or
fixed-base URLs (ssrf); public slug lookups and admin-role routes (authz);
local `./` actions and reusable workflows (unpinned-action); `npm install -g`
(lockfile-rewrite); dummy test passwords and cookie-name constants (secret);
"skip review apps" and "do not report to Sentry" comments (untrusted-diff).

## Limitations

- The corpus is author-written by one person, so it reflects one reviewer's
  judgement and blind spots; it is not a sample of real-world PRs.
- N is small (111 cases, 57 expected findings). Per-class numbers rest on 4 to
  7 positives; one case moves a score by 15 to 25 points. Treat results as
  directional, not as statistically significant.
- Several cases deliberately target known weak spots, so absolute scores are
  pessimistic compared with a natural distribution of PRs.
- Some labels are debatable (for example wildcard CORS on a public asset, or
  first-party Actions policy). They are recorded as the author judged them.
- Only the diff is analysed; cases cannot show code outside the hunks.
