# Skill — Injection: SQL, OS command, path traversal

## Purpose
Find request input or other non-literal values built into a SQL string, a shell command, or a file path in a PR diff. Source → sink, or drop it.

## Risk class
Cite these IDs only. Do not invent CWE / CVE / CVSS / extra catalog numbers.

| Kind | IDs (official) | Why this skill |
|---|---|---|
| OWASP Top 10:2021 | [A03:2021](https://owasp.org/Top10/A03_2021-Injection/) Injection | SQL (CWE-89) and OS command (CWE-78) injection are in the A03 CWE list |
| OWASP Top 10:2021 | [A01:2021](https://owasp.org/Top10/A01_2021-Broken_Access_Control/) Broken Access Control | Path traversal (CWE-22) is in the A01 CWE list, not A03 |
| OWASP ASVS 5.0 | [1.2.4](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x10-V1-Encoding-and-Sanitization.md) parameterized queries, ORMs or equivalent protection against SQL injection; [1.2.5](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x10-V1-Encoding-and-Sanitization.md) protection against OS command injection, parameterized OS calls or contextual command-line encoding | SQL and shell sinks |
| OWASP ASVS 5.0 | [5.3.2](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x14-V5-File-Handling.md) file paths built from internally generated or trusted data, or user-submitted names under strict validation and sanitization | File read / write / send sinks |
| Agentic posture | [ASI02:2026](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/) Tool Misuse and Exploitation; [AST03](https://owasp.org/www-project-agentic-skills-top-10/) Over-Privileged Skills | The reviewer never runs the query, the command or the file read from the diff. Location + why is enough. |

The Go CLI prints, per class:

| Class | CWE | IDs printed |
|---|---|---|
| SQL injection | CWE-89 | `A03:2021, ASVS-5.0-1.2.4` |
| OS command injection | CWE-78 | `A03:2021, ASVS-5.0-1.2.5` |
| Path traversal | CWE-22 | `A01:2021, ASVS-5.0-5.3.2` |

ASI/AST stay here. The JS engine (`npm run review`) has no injection rules; use the Go CLI (`make build`, then `bin/review-go <diff>`).

## Hunt table

| Signal in the diff | What you are looking for | Evidence to cite |
|---|---|---|
| `query` / `execute` / `raw` / `Exec` / `Query` / `QueryRow` / `createQuery` / `executeQuery` / `prepare` / `mysqli_query` / `pg_query` | SQL text built with `+`, `.`, a template literal, an f-string, `.format`, `%` or `fmt.Sprintf` from a non-literal | sink `path:line` + where the value comes from |
| `exec` / `execSync` (`child_process`), `spawn` / `execFile` with `shell: true` | Command string built from a non-literal | call `path:line` |
| `subprocess.*(..., shell=True)`, `os.system`, `os.popen` | Non-literal command | call `path:line` |
| `exec.Command("sh", "-c", x)`, `new ProcessBuilder("sh", "-c", x)`, `Runtime.getRuntime().exec(...)` | Non-literal after `-c`, or a concatenated command | call `path:line` |
| PHP `shell_exec` / `system` / `passthru` / `exec` / `popen` / `proc_open` / backticks | A `$variable` not wrapped in `escapeshellarg` | call `path:line` |
| `fs.readFile` / `writeFile` / `createReadStream` / `res.sendFile` / `res.download`, Python `open` / `send_file`, Go `os.Open` / `os.ReadFile` / `http.ServeFile`, Java `new File` / `Paths.get` / `Files.*`, PHP `file_get_contents` / `fopen` / `include` | Path built from `req.params` / `req.query` / `req.body`, `request.args`, `r.URL.Query()`, `$_GET` / `$_POST`, `@RequestParam` | sink `path:line` + source `path:line` |

Request input that visibly reaches the sink in the diff is **HIGH**. A non-literal SQL or shell value whose source is outside the diff is **MEDIUM** (review). Path traversal only fires on visible request input.

## Anti-patterns (look-alikes)

| In the hunk | Usually **not** a finding | Usually **FAIL** (cite `path:line`) |
|---|---|---|
| `db.query("... WHERE id = $1", [id])`, `cur.execute("... %s", (x,))`, `prepareStatement("... ?")` | Parameterized: value bound, not built | Same query with `+ id` or `${id}` in the SQL text |
| `` `SELECT * FROM ${'users'}` ``, `"... FROM " + USERS_TABLE` | Interpolated value is a literal or an UPPER_CASE constant | Interpolated value is a variable or a call |
| `exec.Command("ping", "-c", "1", host)`, `spawn("ls", [dir])`, `subprocess.run(["ls", d])` | argv form, no shell parses the value | Same value in `sh -c`, `shell: true` or `shell=True` |
| `escapeshellarg($x)` around every variable | Escaped for the shell | Any bare `$x` in the command |
| `path.resolve` + `startsWith(base)`, `realpath` + `commonpath`, `filepath.Clean` + `strings.HasPrefix`, `basename`, `secure_filename`, `sendFile(name, { root })` in the same handler | Visible normalization + base-directory check | `path.join(base, req.params.x)` alone: `join` keeps `..` |

## Required evidence
Every finding: **`path:line`** of the sink. Prefer a second `path:line` for the source.

If the source is outside the diff, say so: MEDIUM at most for SQL and shell, nothing for paths. No `path:line` → drop it. Guardrail: [`guardrails/evidence-required.md`](../guardrails/evidence-required.md).

## Pass / fail
- **FAIL** — a non-literal value is built into SQL text or a shell command, or request input reaches a file path with no normalization + base-directory check in the same handler.
- **PASS** — parameterized query, argv-style process call, escaped shell argument, or a visible path guard. Cite that `path:line` too.
- **INSUFFICIENT** — the sink or the source is guessed, or the value is built outside the diff. Do not invent CWE-89 / CWE-78 / CWE-22.

Argument injection into an argv call (a leading `-` read as an option) is not shell injection. Note it for the human as **insufficient evidence**; do not label it CWE-78.

## Fix direction
1. SQL: bind every value as a parameter. Identifiers (column, table, sort order) cannot be bound: map them through a fixed allowlist.
2. Shell: drop the shell. Call the program with an argv list; validate the value; end options with `--` where the tool supports it. If a shell is unavoidable, escape every argument for that shell.
3. Paths: prefer an internal id → path map. Otherwise resolve the path and require it to stay under the base directory, or keep only the basename.
4. Least-privilege database users, sandboxes and read-only mounts are defense in depth, not the fix for this PR.

## Out of scope
- Writing payloads, exploit proof-of-concepts or "try this input".
- Claiming which rows, files or hosts are reachable without schema, grants or deployment in the diff.
- NoSQL, LDAP, XPath, template or header injection — other sinks, not this skill.
- Running the query, the command or the file read to "confirm". The reviewer reads; it does not execute.

## Output format
- `SEVERITY` | `file:line` | summary | (optional CWE)
- One line on source → sink, or `insufficient evidence`.

## Refs (public)
- [OWASP Top 10:2021 A03 Injection](https://owasp.org/Top10/A03_2021-Injection/)
- [OWASP Top 10:2021 A01 Broken Access Control](https://owasp.org/Top10/A01_2021-Broken_Access_Control/) (lists CWE-22)
- [OWASP ASVS 5.0 V1 Encoding and Sanitization](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x10-V1-Encoding-and-Sanitization.md) (`1.2.4`, `1.2.5`)
- [OWASP ASVS 5.0 V5 File Handling](https://github.com/OWASP/ASVS/blob/v5.0.0/5.0/en/0x14-V5-File-Handling.md) (`5.3.2`)
- [OWASP SQL Injection Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html)
- [OWASP OS Command Injection Defense Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html)
- [CWE-89](https://cwe.mitre.org/data/definitions/89.html), [CWE-78](https://cwe.mitre.org/data/definitions/78.html), [CWE-22](https://cwe.mitre.org/data/definitions/22.html) — only when the sink is in the hunk

Worked fixtures: [`examples/sql-injection.sample.diff`](../examples/sql-injection.sample.diff) → [`examples/sql-injection.sample-report.md`](../examples/sql-injection.sample-report.md); [`examples/command-injection.sample.diff`](../examples/command-injection.sample.diff) → [`examples/command-injection.sample-report.md`](../examples/command-injection.sample-report.md); [`examples/path-traversal.sample.diff`](../examples/path-traversal.sample.diff) → [`examples/path-traversal.sample-report.md`](../examples/path-traversal.sample-report.md). Index: [`examples/README.md`](../examples/README.md).
