# Third blind set: five language families, nine classes

80 labelled PR diffs written on 2026-11-12. The second blind set
(`../blind2`) stopped being blind once its gaps were fixed, so this set is the
next unbiased measurement. Same layout and `expected.json` format as
[`../README.md`](../README.md).

```
go run ./cmd/bench --corpus bench/blind3 --json bench/blind3-results.json
```

## Scope

- 16 cases each for Python, Go, Java/Kotlin, PHP and JS/TS.
- Classes: authz 10, xss 11, cors 9, ssrf 9, sql-injection 9,
  command-injection 9, path-traversal 9, secret 8, supply chain 6
  (install script, lockfile rewrite, registry host, unpinned action, and two
  clean cases).
- 42 positives and 38 hard negatives: owner-scoped queries, allowlists,
  `escapeshellarg`, `htmlspecialchars`, `html.EscapeString`, basename and
  normalize checks, parameterized queries, `${ENV}` placeholders, SHA-pinned
  actions, a regenerated lockfile with identical hashes, a minified bundle, a
  test file and a plain-text mail template.
- Frameworks: Flask, Django, FastAPI, net/http, chi, gin, echo, fiber, Spring
  MVC and WebFlux, Ktor, JAX-RS, Servlet, Micronaut, plain PHP, Laravel,
  Symfony, Slim, Express, Fastify, NestJS, Next.js, Koa, React and Vue.

## How it was written

The same protocol as the first two blind sets:

- The author could not read `internal/rules/`, `internal/review/`,
  `bench/corpus/`, `bench/blind/`, `bench/blind2/`, `examples/`, `testdata/`,
  `security/examples/`, `skills/`, `docs/benchmark.md` or any test file.
- The engine was not run while writing, so no label was matched to its output.
- One expected finding per positive: the added line of the sink, by new-file
  line number.
- A separate parser confirmed hunk headers and that every expected `path:line`
  is an added line.
- The only edits after scoring started: the author wrote the class names `cmdi`
  and `sqli`, which the bench rejects; they were renamed to `command-injection`
  and `sql-injection`. Three fake keys (two Stripe-shaped, one SendGrid-shaped)
  were shortened so GitHub push protection accepts the files; the scores were
  the same before and after. No line or verdict was changed.

## Keeping it blind

The first run in [`docs/benchmark.md`](../../docs/benchmark.md) stays the
honest number. Fixes for the gaps it shows must be checked against new cases in
`bench/corpus`; once a rule changes while looking at a case here, the set is no
longer blind.
