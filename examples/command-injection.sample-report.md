# Sample report — OS command injection (fixture)

Not a production incident. Walkthrough of [`skills/injection.md`](../skills/injection.md) against [`command-injection.sample.diff`](command-injection.sample.diff).

## Hunt
- Diff adds `Ping`: the `host` query parameter is concatenated into the command string of `exec.Command("sh", "-c", ...)` (hunt table: shell with a non-literal command).
- Diff adds `Trace`: the same input goes to `exec.Command("traceroute", ...)` as one argv element. No shell parses it — anti-pattern table: **not** an OS command injection finding.
- `Uptime` runs a literal command → not a finding.

## Findings

```text
HIGH | internal/tools/ping.go:15 | r.URL.Query() value concatenated into a sh -c command string | CWE-78
```

## Evidence
- Source: `internal/tools/ping.go:14` — `host := r.URL.Query().Get("host")`.
- Sink: `internal/tools/ping.go:15` — `exec.Command("sh", "-c", "ping -c 1 "+host)`.
- Explicit non-finding: `internal/tools/ping.go:25` — argv form without a shell.
- Explicit withhold: whether `traceroute` treats a leading `-` in `host` as an option is argument injection, not shell injection; the rule does not claim it. Note it for the human reviewer as **insufficient evidence**.

## Verdict
**FAIL**

## Fix direction
Drop the shell at `internal/tools/ping.go:15`: call `exec.Command("ping", "-c", "1", host)`. Validate `host` as a hostname or IP before use, and end options with `--` where the tool supports it.

## Out of scope / not claimed
- No claim about the container, its user or what a shell could reach.
- No CVSS, no exploit PoC, no payload.
- Agent does not patch the handler. Human decides request-changes.
