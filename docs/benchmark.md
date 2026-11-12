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

Several numbers, kept apart on purpose. Only the first run of each blind set
estimates accuracy on code the rules have not seen.

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
these numbers are **fitted, not an accuracy estimate**. It now has 382 cases, including the 15 `*-n*` cases from the real-PR noise fixes (section 7), the four `untrusted-d*` docs cases written with the docs-noise fix, the `*-g*` cases written for the gap fixes below and the 21 `cf-*` Cloudflare cases. The `cf-*` cases were written together with the Cloudflare rules, so they show the rules work as designed, not how well they generalise. The same holds for the Python, Go, Java and PHP cases of the web rules (`*-l*`, `*-j*`, `*-p*`, `*-d*`, `*-f*`, `*-c*`), written with the language extensions. Before those merged, a separate reviewer ran 45 idiomatic safe diffs and 16 vulnerable ones in those languages, outside the corpus. It found four misses (two AuthZ false positives, a PHP cast false positive and a servlet recall gap), fixed before merge. After the fixes, all 45 safe diffs stay silent. The blind set below is unchanged by the extensions (47 TP, 6 FP, 9 FN).

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 20 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cloudflare-config | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| command-injection | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cors | 20 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| install-script | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| lockfile-rewrite | 9 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| path-traversal | 3 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| registry-host | 24 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| secret | 31 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| sql-injection | 3 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| ssrf | 20 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| unpinned-action | 7 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| untrusted-diff | 17 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| xss | 31 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| **overall** | 207 | 0 | 0 | 1.00 | 1.00 | 1.00 |

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

### 5. Second blind set (`bench/blind2`), Python, Go, Java and PHP

A fresh set of 64 cases written on 2026-11-12 under the same protocol, after
the web rules gained four languages. See
[`bench/blind2/README.md`](../bench/blind2/README.md). First run, with no rule
changes made against it:

```
go run ./cmd/bench --corpus bench/blind2 --json bench/blind2-results.json
```

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 4 | 1 | 2 | 0.80 | 0.67 | 0.73 |
| command-injection | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cors | 4 | 0 | 1 | 1.00 | 0.80 | 0.89 |
| path-traversal | 4 | 1 | 1 | 0.80 | 0.80 | 0.80 |
| sql-injection | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| ssrf | 5 | 1 | 0 | 0.83 | 1.00 | 0.91 |
| xss | 5 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| **overall** | 32 | 3 | 4 | 0.91 | 0.89 | 0.90 |

64 cases (36 expected findings, 28 clean cases of which 27 stayed silent).
**This is the current accuracy estimate for the four new languages.** Open gaps:

- AuthZ on JAX-RS and FastAPI delete handlers is missed (2 FN); an echo download
  handler that keeps only the base name is flagged as AuthZ (1 FP).
- Flask reflected origin with credentials is missed (1 FN).
- Spring `Paths.resolve` without a containment check is missed (1 FN).
- PHP `readfile` and `file_get_contents` on request input are reported once as
  SSRF and once as path traversal, the other class from the label (1 FP each).
  Both readings are defensible; the line is right.

After the gap fixes (made while looking at these diffs, so no longer blind):
33 TP, 3 FP, 3 FN, F1 0.92. The fixes were written from one-line gap
descriptions and checked on new corpus cases. Against these diffs they closed
the Flask CORS miss and the echo AuthZ false positive. The Spring
`Paths.resolve` case is now flagged one line below its label. The JAX-RS and
FastAPI delete handlers are still missed: a rule fitted to its own corpus case
did not carry over.

### 6. Triage pass on the blind sets

The optional triage pass (`--triage`, role in
[`reviewers/triage.md`](../reviewers/triage.md)) was run with a real model on
every finding of both blind sets. The model saw only the role, the guardrail,
each diff and the engine report; it never saw a label. Its verdicts are
committed in `bench/triage/`, and the score is reproducible:

```
go build -o /tmp/review ./cmd/review
python3 -I scripts/score-triage.py /tmp/review bench/blind bench/triage/blind
python3 -I scripts/score-triage.py /tmp/review bench/blind2 bench/triage/blind2
```

| Set | Findings | TP | FP | FP dropped | TP dropped or downgraded | Precision before | Precision after |
|---|---|---|---|---|---|---|---|
| `bench/blind` | 53 | 47 | 6 | 2 | 0 | 0.89 | 0.92 |
| `bench/blind2` | 35 | 32 | 3 | 1 | 0 | 0.91 | 0.94 |
| **both** | 88 | 79 | 9 | 3 | 0 | 0.90 | 0.93 |

The pass removed a third of the false positives and no true positive. It stayed
silent on the other six, which matches the role: doubt keeps the finding. It
cannot help recall, since a verdict can only lower. One model run per set; a
different model or run may emit different verdicts.

### 7. Real pull requests (noise)

The blind sets are author-written. To see how the engine behaves on real
work, it reviewed the 40 most recent merged pull requests that change
application code in six public projects, 240 PRs in all, on 2026-11-12:
paperless-ngx (Python, Django), gitea (Go), halo (Java, Spring), invoiceninja
(PHP, Laravel), cal.com (TypeScript, Next.js) and outline (TypeScript, Koa and
React). A reviewer read every finding's hunk and labelled it. The list of PRs,
every finding with its label, and the triage verdicts are in
[`bench/real-prs/`](../bench/real-prs/); the diffs themselves are not copied.
There is no ground truth for what these PRs should have flagged, so this
measures noise, not recall.

| Run | PRs blocked | Findings | Real issues among the blocks |
|---|---|---|---|
| Engine before the fixes | 9 of 240 (3.8%) | 65 | 0 (59 wrong, 6 debatable) |
| Engine before the fixes, plus triage | 3 of 240 (1.3%) | 4 | 0 |
| Engine after the fixes | 3 of 240 (1.3%) | 3 | 0 (1 wrong, 2 debatable) |

The first run blocked nine PRs and none was a real issue. The noise came from
four shapes, fixed in the engine and pinned by new `*-n*` corpus cases:

- **Lockfile reformat (39 findings).** npm re-emitted entries with a trailing
  comma after an unchanged `integrity`; the rule now compares the hash value.
- **Minified bundle (7).** A hashed `public/build/assets/app-*.js` with a
  100 KB line. Code-sink rules now skip generated output (`*.min.*`, source maps,
  `dist/`, `.next/`, `public/build/`, hashed names under `build/`) and lines over
  2000 characters. Secrets and untrusted-diff still read them.
- **Plain-text email views (14).** `{!! !!}` in `*_text.blade.php` renders
  text/plain, not HTML; `{!! ctrans('literal') !!}` is static.
- **Test code (3).** Code-sink rules skip test files; in test files a
  `sk_test_`/`pk_test_` key, or a token with a short tail, is a placeholder,
  while a live-looking key still fires.

The fixes were written while looking at these PRs, so the "after" row is
fitted to them; the next sample of real PRs is the honest check. The blind sets
did not move (sections 5 and 6 still hold). The three findings left are an
`innerHTML` from an icon map, an `innerHTML` from the app's own endpoint, and a
PEM key in a test helper. The triage pass, run on the first report, dropped 61
of the 65 findings and downgraded one; the three blocks it left are the
test-file secrets it may not touch by design.

### 8. Real pull requests, second sample (out of sample)

The fixes in section 7 were fitted to their PRs, so a second sample was taken
on 2026-11-12 from eight projects the engine had never seen: netbox and zulip
(Python), grafana and gogs (Go), keycloak (Java), matomo (PHP), Ghost and n8n
(JS/TS). The 40 most recent merged PRs that change application code in each,
315 reviewed (5 skipped: diffs over 2 MB or rejected by the GitHub API). Same
labelling as section 7: one reviewer per finding, from the hunk. Data in
[`bench/real-prs2/`](../bench/real-prs2/).

| Run | PRs blocked | Findings | Real issues |
|---|---|---|---|
| Engine at section 7, first run | 6 of 315 (1.9%) | 15 | 0 (10 wrong, 5 debatable) |
| After the fixes below (fitted) | 3 of 315 (1.0%) | 6 | 0 (1 wrong, 5 debatable) |

Half the blocks of the first sample's starting point, on code the fixes never
saw, but every block is still wrong. Four projects (netbox, zulip, keycloak,
Ghost) produced no finding at all; matomo produced 11 of the 15. The wrong
ones come from five shapes, still open here:

- **SQL text read as a shell command (5).** A multi-line PHP SQL string with
  `` FROM `$table` `` and `%1$s` placeholders passed to a bound-query helper.
- **Constant SQL concatenation (2).** `'INSERT INTO ' . prefixTable(CONST)`
  and `'IN (' . placeholderList($ids) . ')'`, with every value bound.
- **Escaped Go HTML (1).** `template.HTML(buf)` after each piece went through
  `html.EscapeString`.
- **AuthZ on an API client (1).** `client.Get().Namespace(ns).Name(x)` forwards
  the caller's token; the server enforces access.
- **A translation string (1).** An Italian UI sentence about AI tools in a
  `lang/*.json` read as an injected instruction.

The second row comes from fixes made while looking at these PRs, so it is
fitted, like section 7's. Each of the five shapes is now pinned by `*-o*` corpus
cases. One SQL line stays: `'UPDATE ' . $this->getTable() . ' ... IN (' .
placeholders . ')'`, where the table name comes from an object method. The
other five findings are the debatable Twig and Vue ones.

No real issue surfaced in 555 real PRs across both samples, so recall on real
code is unmeasured; the blind sets are the recall estimate.

### 9. Third blind set (`bench/blind3`), five language families

80 cases written on 2026-11-12 under the same protocol, across Python, Go,
Java/Kotlin, PHP and JS/TS, with secrets and supply chain added. See
[`bench/blind3/README.md`](../bench/blind3/README.md). Scored on the engine
after the section 5 gap fixes, which were made without seeing this set:

```
go run ./cmd/bench --corpus bench/blind3 --json bench/blind3-results.json
```

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 1 | 0 | 4 | 1.00 | 0.20 | 0.33 |
| command-injection | 4 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| cors | 4 | 0 | 1 | 1.00 | 0.80 | 0.89 |
| install-script | 1 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| lockfile-rewrite | 0 | 0 | 1 | n/a | 0.00 | n/a |
| path-traversal | 3 | 1 | 1 | 0.75 | 0.75 | 0.75 |
| registry-host | 1 | 1 | 0 | 0.50 | 1.00 | 0.67 |
| secret | 4 | 2 | 1 | 0.67 | 0.80 | 0.73 |
| sql-injection | 4 | 1 | 1 | 0.80 | 0.80 | 0.80 |
| ssrf | 5 | 1 | 0 | 0.83 | 1.00 | 0.91 |
| unpinned-action | 1 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| xss | 4 | 5 | 1 | 0.44 | 0.80 | 0.57 |
| **overall** | 32 | 11 | 10 | 0.74 | 0.76 | 0.75 |

80 cases (42 expected findings, 38 clean cases of which 35 stayed silent).
**This is the current out-of-sample estimate, and it is lower than the second
set's 0.90 F1.** The strict matcher counts a finding as wrong unless path, line
and class all match; read case by case:

- **Missed (6 real gaps).** AuthZ on a by-id lookup in chi, Spring
  (`findById`), Laravel (`find`) and Flask (4); fiber wildcard origin with
  credentials (1); JAX-RS `new File` on a path parameter (1). Also missed: an
  echo handler concatenating input into HTML, and a SendGrid key in
  `.env.production`.
- **Wrong (4).** XSS on `htmlspecialchars` output; XSS on `echo shell_exec`
  of an `escapeshellarg` argument; two `{!! !!}` lines in a
  `*-text.blade.php` mail view (the plain-text name list knows `_text` and
  `.text`, not `-text`).
- **Right line, other reading (5).** A second real secret on the next line of
  the same file (2); PHP `readfile` and `file_get_contents` reported as the
  other of SSRF and path traversal (2); `echo shell_exec($_GET…)` also reported
  as XSS (1).
- **Right issue, other label (2 FN + 2 FP).** SQL built with `fmt.Sprintf`
  flagged on the query call one line below the label; a changed lockfile hash
  reported as registry host instead of lockfile rewrite.

After fixing these gaps (written from one-line descriptions, checked on new
`*-o*` corpus cases, and only then re-run here, so no longer blind): **41 TP,
5 FP, 1 FN, precision 0.89, recall 0.98, F1 0.93.** The AuthZ misses had one
cause: the rules required the route or handler line to be added, while a real
PR usually adds only the lookup inside an unchanged handler. Every remaining
error is in the "right line, other reading" or "other label" groups above. The
other two sets moved too: `bench/blind` to 48/5/8 (F1 0.88, from the lockfile
class fix), and `bench/blind2` stayed at 33/3/3.

### 10. Fourth blind set (`bench/blind4`), less common frameworks

80 cases written on 2026-11-12 after the section 9 fixes, on frameworks the
earlier sets did not use (aiohttp, Tornado, Starlette, fasthttp, gorilla/mux,
Vert.x, Ktor, Quarkus, CodeIgniter 4, Laminas, Hono, Remix and others), with
most sinks added inside an unchanged handler. See
[`bench/blind4/README.md`](../bench/blind4/README.md). First run, engine at
#48, no rule changed against it:

```
go run ./cmd/bench --corpus bench/blind4 --json bench/blind4-results.json
```

| Class | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| authz | 3 | 0 | 2 | 1.00 | 0.60 | 0.75 |
| command-injection | 4 | 0 | 1 | 1.00 | 0.80 | 0.89 |
| cors | 3 | 0 | 1 | 1.00 | 0.75 | 0.86 |
| install-script | 1 | 1 | 0 | 0.50 | 1.00 | 0.67 |
| lockfile-rewrite | 0 | 0 | 1 | n/a | 0.00 | n/a |
| path-traversal | 0 | 0 | 4 | n/a | 0.00 | n/a |
| registry-host | 1 | 1 | 0 | 0.50 | 1.00 | 0.67 |
| secret | 4 | 0 | 1 | 1.00 | 0.80 | 0.89 |
| sql-injection | 4 | 0 | 1 | 1.00 | 0.80 | 0.89 |
| ssrf | 2 | 1 | 3 | 0.67 | 0.40 | 0.50 |
| unpinned-action | 0 | 0 | 1 | n/a | 0.00 | n/a |
| xss | 3 | 0 | 2 | 1.00 | 0.60 | 0.75 |
| **overall** | 25 | 3 | 17 | 0.89 | 0.60 | 0.71 |

80 cases (42 expected findings, 38 clean cases of which 36 stayed silent).
**Precision holds out of sample; recall does not.** The engine is quiet and
mostly right when it speaks, but on frameworks it has no patterns for it
says nothing:

- **No source or sink pattern for the framework (13 FN).** Path traversal in
  aiohttp, Quarkus, CodeIgniter 4 and a Go static handler with `filepath.Join`
  (4); SSRF in aiohttp, Ktor and CodeIgniter curl (3); XSS in fasthttp and
  Vert.x (2); Ktor `anyHost()` CORS (1); Spring `JdbcTemplate` string query (1);
  a Java shell `ProcessBuilder("sh", "-c", ...)` (1); a password in a Laminas
  PHP config array (1).
- **AuthZ (2 FN).** A gorilla/mux and a Spring by-id lookup still pass.
- **Label convention (2).** A changed `resolved` host in a lockfile is
  `registry-host` for the engine and `lockfile-rewrite` for the author (1 FN +
  1 FP). `actions/setup-node@v4` on a tag is exempt by design: the rule trusts
  the `actions/` and `github/` owners (1 FN).
- **Wrong (2 FP).** SSRF on a Go proxy that parses the URL and compares
  `Hostname()` to one fixed host before the request; an install-script finding
  on a `postinstall: patch-package` the author called benign (any install hook
  fires by design).

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
