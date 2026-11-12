# Sample report — XSS in server-rendered HTML (fixture)

Not a production incident. Walkthrough of [`security/skills/xss-html.md`](../skills/xss-html.md) against [`xss-server-html.sample.diff`](xss-server-html.sample.diff). Shows the server-side languages (Python, Go; Java and PHP follow the same shape).

## Hunt
- `app/search.py`: the view returns an f-string HTML page; `q` comes from `request.args`.
- `internal/web/hello.go`: the handler writes HTML with `fmt.Fprintf`, but the query value goes through `html.EscapeString` first.

## Findings
- `HIGH` | `app/search.py:10` | `request.args` value `q` interpolated into an HTML string returned by the view with no escape in the hunk | CWE-79

## Evidence
- Sink: `app/search.py:10` — `return f"<h1>Results for {q}</h1>"`.
- Source in the same file: `app/search.py:9` — `request.args.get("q", "")`.
- `escape` is imported at `app/search.py:2` but never applied to `q`.
- Not a finding: `internal/web/hello.go:11` — `html.EscapeString(name)` encodes the value for the HTML body before the sink.

## Verdict
**FAIL**

## Fix direction
Wrap the value with `escape(q)` or render a Jinja template, which auto-escapes. Do not regex-strip `<script>`.

## Out of scope / not claimed
- No exploit payload, no CVSS, no claim about stored vs reflected beyond the diff.
