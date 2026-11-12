# Fourth blind set: less common frameworks, edits inside existing files

80 labelled PR diffs written on 2026-11-12, after the fixes that followed the
third blind set. Same layout and `expected.json` format as
[`../README.md`](../README.md).

```
go run ./cmd/bench --corpus bench/blind4 --json bench/blind4-results.json
```

## Scope

- 16 cases each for Python, Go, Java/Kotlin, PHP and JS/TS; 42 positives and
  38 hard negatives.
- Classes: authz 10, xss 10, cors 8, ssrf 9, sql-injection 9,
  command-injection 9, path-traversal 9, secret 8, and two cases each for
  install-script, lockfile-rewrite, registry-host and unpinned-action.
- Frameworks chosen away from the obvious ones: Django REST Framework,
  Starlette, aiohttp, Tornado, gorilla/mux, fasthttp, `ServeMux` with
  `PathValue`, Quarkus, Vert.x, Ktor, Spring Data JPA, CodeIgniter 4, CakePHP,
  WordPress, Laminas, Hapi, Hono, Remix, SvelteKit and Angular.
- 75 cases edit an existing file; in about 51 the handler, route or signature
  is an unchanged context line and only the sink is added.

## How it was written

The same protocol as the earlier blind sets: the author could not read the
engine, its tests, the corpus, the other blind sets, the real-PR samples or the
docs, and never ran the engine. One expected finding per positive, on the sink
line, by new-file line number. A separate parser checked hunk headers, that
every expected `path:line` is an added line, and the class names. Fake keys
were written below provider lengths so push protection accepts them. No file
was edited after scoring.

## Keeping it blind

The first run in [`docs/benchmark.md`](../../docs/benchmark.md) stays the
honest number. Fixes for its gaps must be checked against new cases in
`bench/corpus`.
