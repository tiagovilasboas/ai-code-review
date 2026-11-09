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

Four numbers, kept apart on purpose. Only the first blind run estimates accuracy on
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
these numbers are **fitted, not an accuracy estimate**. It now has 304 cases, including the `*-g*` cases written for the gap fixes below.

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 15 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| command-injection | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cors | 16 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| install-script | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| lockfile-rewrite | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| path-traversal | 3 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| registry-host | 23 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| secret | 25 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| sql-injection | 3 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| ssrf | 16 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| unpinned-action | 7 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| untrusted-diff | 16 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| xss | 19 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| **overall** | 165 | 0 | 0 | 1.00 | 1.00 | 1.00 |

### 3. Blind set (`bench/blind`)

Written by a separate author who could not read the rules, the corpus or the
examples, and never ran the tool before labelling. See
[`bench/blind/README.md`](../bench/blind/README.md). First run on 2026-11-09,
with no rule changes made against it:

```
go run ./cmd/bench --corpus bench/blind --json bench/blind-results.json
```

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 1 | 0 | 3 | 1.00 | 0.25 | 0.40 |
| command-injection | 4 | 1 | 1 | 0.80 | 0.80 | 0.80 |
| cors | 1 | 1 | 3 | 0.50 | 0.25 | 0.33 |
| install-script | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| lockfile-rewrite | 0 | 0 | 4 | n/a | 0.00 | n/a |
| path-traversal | 5 | 2 | 0 | 0.71 | 1.00 | 0.83 |
| registry-host | 2 | 3 | 2 | 0.40 | 0.50 | 0.44 |
| secret | 0 | 2 | 4 | 0.00 | 0.00 | 0.00 |
| sql-injection | 4 | 1 | 1 | 0.80 | 0.80 | 0.80 |
| ssrf | 2 | 0 | 2 | 1.00 | 0.50 | 0.67 |
| unpinned-action | 4 | 1 | 0 | 0.80 | 1.00 | 0.89 |
| untrusted-diff | 3 | 0 | 1 | 1.00 | 0.75 | 0.86 |
| xss | 3 | 0 | 2 | 1.00 | 0.60 | 0.75 |
| **overall** | 33 | 11 | 23 | 0.75 | 0.59 | 0.66 |

109 cases (56 expected findings, 53 clean cases of which 46 stayed silent).

**This is the accuracy estimate.** It is close to the pre-tuning baseline,
which means the precision work fixed the cases it was aimed at but generalised
less than the tuned numbers suggest. Main gaps on unseen code:

- Lockfile rewrites to a new host are reported as registry-host instead of
  lockfile-rewrite (4 FN plus 3 FP of the wrong class).
- Secrets in Django settings, Spring `.properties`, Go DSNs and JWT signing
  literals are missed; Secret Manager resource names are flagged.
- AuthZ on gin, NestJS and Flask handlers is missed; reflected-origin CORS in
  Go, Koa and Spring is missed.
- Allowlist and containment checks (`filepath.Rel`, template-name and
  sort-column allowlists, `escapeshellarg`) are not recognised as safe.

To keep the set blind, fixes for these gaps must be checked against new cases,
not against these. Once a rule is changed while looking at a blind case, that
case moves to `bench/corpus`.

### 4. Blind set after the gap fixes (no longer blind)

After the first blind run, five fixes closed the gaps it showed (PRs #25 to #29).
The fixes were written from short descriptions of each gap and checked against
new `*-g*` cases in `bench/corpus`; nobody read the blind diffs while fixing. But
the gaps themselves came from this set, so **this run is informed by it and is
not an unbiased estimate**. Section 3 remains the honest first measurement; a
fresh blind set is needed for the next one.

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| command-injection | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cors | 4 | 1 | 0 | 0.80 | 1.00 | 0.89 |
| install-script | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| lockfile-rewrite | 0 | 0 | 4 | n/a | 0.00 | n/a |
| path-traversal | 5 | 1 | 0 | 0.83 | 1.00 | 0.91 |
| registry-host | 4 | 4 | 0 | 0.50 | 1.00 | 0.67 |
| secret | 0 | 0 | 4 | n/a | 0.00 | n/a |
| sql-injection | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| ssrf | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| unpinned-action | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| untrusted-diff | 3 | 0 | 1 | 1.00 | 0.75 | 0.86 |
| xss | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| **overall** | 47 | 6 | 9 | 0.89 | 0.84 | 0.86 |

109 cases (56 expected findings, 53 clean cases of which 51 stayed silent).
What is left, and why:

- **lockfile-rewrite (4 FN, 4 FP of the wrong class):** a taxonomy mismatch,
  not a missed detection. In this project `lockfile-rewrite` is a CI step that
  rewrites the lockfile; a lockfile entry pointing at a new host or hash is
  `registry-host`. The engine flags the right line under that class. Counting
  the line regardless of class, these four are hits.
- **secret (4 FN):** every fake value in these cases contains `EXAMPLE` or
  `0000`, or is the AWS documentation example key. The engine treats those as
  documentation placeholders on purpose, so it stays silent. The same shapes
  with non-placeholder values are covered by `secret-g*` cases in the corpus.
- **untrusted-diff (1 FN), cors (1 FP), path-traversal (1 FP):** still open. The
  CORS case (wildcard on a public, credential-less status route) is a debatable
  label.

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
