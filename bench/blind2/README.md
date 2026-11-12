# Second blind set: Python, Go, Java and PHP

64 labelled PR diffs written on 2026-11-12, after the web rules were extended
to four more languages. The first blind set (`../blind`) stopped being blind
once its gaps were fixed, so this set is the next unbiased measurement. Same
layout and `expected.json` format as [`../README.md`](../README.md).

```
go run ./cmd/bench --corpus bench/blind2 --json bench/blind2-results.json
```

## Scope

- 16 cases per language: Python, Go, Java and PHP.
- Classes: authz, xss, cors, ssrf, sql-injection, command-injection,
  path-traversal (9 or 10 cases each).
- 36 positives and 28 hard negatives: owner-scoped queries, `@PreAuthorize`
  in the hunk, parameterized queries, constant base URLs, argv-only exec,
  basename and normalize checks, auto-escaped templates, CORS allowlists.
- Frameworks: Flask, Django, FastAPI, net/http, chi, gin, echo, Spring MVC,
  Spring WebFlux, JAX-RS, Servlet, JSP, Thymeleaf, Micronaut, plain PHP,
  Laravel, Symfony and Slim.

## How it was written

The same protocol as the first blind set:

- The author could not read `internal/rules/`, `internal/review/`,
  `bench/corpus/`, `bench/blind/`, `examples/`, `testdata/`,
  `security/examples/`, `skills/`, `docs/benchmark.md` or any test file.
- The engine was not run while writing, so no label was matched to its output.
- One expected finding is the added line of the sink, by new-file line number.
  When the source or sink is outside the diff, the label is silence.
- A separate format check confirmed hunk headers and that every expected
  `path:line` is an added line.

## Keeping it blind

Fixes for the gaps this set shows must be checked against new cases in
`bench/corpus`, not against these diffs. Once a rule is changed while looking
at a case here, the set is no longer blind and the first run in
[`docs/benchmark.md`](../../docs/benchmark.md) stays the honest number.
