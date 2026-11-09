# Benchmark corpus

Labelled diffs used by `make bench` (`go run ./cmd/bench`) to measure the
review engine. Method, results and limitations: [docs/benchmark.md](../docs/benchmark.md).

```
bench/
  corpus/<case-id>/change.diff     unified diff fed to the engine
  corpus/<case-id>/expected.json   what a careful AppSec reviewer would flag
  results.json                     last run, written by `make bench`
```

`expected.json`:

```json
{
  "description": "what the case shows and why it is (not) a finding",
  "source": "author-written",
  "findings": [{ "path": "src/routes/orders.js", "line": 6, "class": "authz" }]
}
```

- `line` is the new-file line number of the flagged line.
- `class` is one of: authz, secret, cors, xss, ssrf, install-script,
  registry-host, unpinned-action, lockfile-rewrite, untrusted-diff,
  sql-injection, command-injection, path-traversal.
- Empty `findings` means the engine must stay silent on that diff.
- Labels reflect reviewer judgement, never the engine's current output.
