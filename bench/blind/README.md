# Blind evaluation set

Labelled PR diffs that measure the review engine on cases its rules were
never tuned against. Same layout and `expected.json` format as
[`../README.md`](../README.md):

```
bench/blind/<case-id>/change.diff     unified diff fed to the engine
bench/blind/<case-id>/expected.json   what a careful AppSec reviewer would flag
```

Classes covered: authz, secret, cors, xss, ssrf, install-script,
registry-host, unpinned-action, lockfile-rewrite, untrusted-diff,
sql-injection, command-injection, path-traversal. Every class has at least
7 cases, about half positives and half hard negatives (sanitized,
parameterized, ownership-checked, allowlisted, pinned, literal-only,
comment-only, test fixtures, look-alike but safe code).

## How it was written

- Written from scratch by an author who did not read the detection logic.
  Off limits while writing: `internal/rules/`, `src/rules/`,
  `internal/review/`, `bench/corpus/`, `examples/`, `testdata/`, `skills/`
  and `docs/benchmark.md`. The only sources consulted were `bench/README.md`
  (format), `internal/finding/` (class names) and `docs/design/cli-contract.md`.
- The engine was not run on these diffs while they were written, so no label
  was adjusted to match its output.
- Labels follow reviewer judgement on the diff alone: flag only what the
  diff shows; when the source or sink is outside the diff, expect silence.
- One expected finding is the single added (`+`) line where the vulnerable
  sink or value appears, by new-file line number.
- Code mixes frameworks and languages (Express, Fastify, Koa, NestJS,
  Next.js, React, Vue, Angular, Svelte, Flask, Django, FastAPI, Go net/http
  and gin, Spring, Laravel and plain PHP, GitHub Actions, npm, yarn and pnpm
  manifests and lockfiles).
- Secret values are fake on purpose (they contain `EXAMPLE` or `0000`, or are
  documented example keys). Host names use reserved example domains.
- A format check confirmed that every diff has consistent hunk headers and
  that every expected `path:line` points at an added line.

## Rules of use

- **Never tune rules against this set.** Read results; do not open a case to
  make a rule pass it.
- If a rule is changed because of a blind case, that case is burned: move it
  to `bench/corpus/` and replace it here with a new case written under the
  same independence rules.
- New blind cases must be written by someone who has not read the rule code
  for that class.
