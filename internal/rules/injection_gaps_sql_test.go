package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestSQLAllowlistedIdentifierIsSilent(t *testing.T) {
	runRuleCases(t, SQLInjection, finding.SQLInjection, "CWE-89", []ruleCase{
		{"ternary allowlist", mkFile("src/api/list.js",
			`+  const col = ALLOWED.includes(req.query.sort) ? req.query.sort : 'id';`,
			"+  const rows = await db.query(`SELECT * FROM users ORDER BY ${col}`);",
		), nil, finding.Medium},
		{"set has reject", mkFile("src/api/list.ts",
			`+  const sort = req.query.sort;`,
			`+  if (!SORTABLE.has(sort)) return res.status(400).end();`,
			`+  const rows = await db.query("SELECT * FROM users ORDER BY " + sort);`,
		), nil, finding.Medium},
		{"map lookup with default", mkFile("src/api/list.js",
			`+  const col = COLUMNS[req.query.sort] ?? 'id';`,
			"+  const rows = await db.query(`SELECT * FROM users ORDER BY ${col}`);",
		), nil, finding.Medium},
		{"inline map lookup", mkFile("src/api/list.js",
			"+  const rows = await db.query(`SELECT * FROM users ORDER BY ${COLUMNS[req.query.sort] || 'id'}`);",
		), nil, finding.Medium},
		{"built var after allowlist", mkFile("src/api/list.js",
			`+  const col = ALLOWED.includes(req.query.sort) ? req.query.sort : 'id';`,
			`+  const sql = "SELECT * FROM users ORDER BY " + col;`,
			`+  const rows = await db.query(sql);`,
		), nil, finding.Medium},
		{"python not in reject", mkFile("app/list.py",
			`+    sort = request.args["sort"]`,
			`+    if sort not in SORTABLE:`,
			`+        abort(400)`,
			`+    cur.execute(f"SELECT * FROM users ORDER BY {sort}")`,
		), nil, finding.Medium},
	})
}

func TestSQLAllowlistKeepsUnsafeVariants(t *testing.T) {
	runRuleCases(t, SQLInjection, finding.SQLInjection, "CWE-89", []ruleCase{
		{"no allowlist", mkFile("src/api/list.js",
			"+  const rows = await db.query(`SELECT * FROM users ORDER BY ${req.query.sort}`);",
		), []int{1}, finding.High},
		{"allowlist guards another name", mkFile("src/api/list.js",
			`+  const col = ALLOWED.includes(req.query.dir) ? req.query.dir : 'asc';`,
			"+  const rows = await db.query(`SELECT * FROM users ORDER BY ${req.query.sort} ${col}`);",
		), []int{2}, finding.High},
		{"check without reject", mkFile("src/api/list.js",
			`+  const sort = req.query.sort;`,
			`+  if (!SORTABLE.has(sort)) console.log("odd sort");`,
			`+  const rows = await db.query("SELECT * FROM users ORDER BY " + sort);`,
		), []int{3}, finding.High},
		{"allowlisted column but raw name", mkFile("src/api/list.js",
			`+  const col = ALLOWED.includes(req.query.sort) ? req.query.sort : 'id';`,
			"+  const rows = await db.query(`SELECT * FROM users WHERE name = '${req.query.name}' ORDER BY ${col}`);",
		), []int{2}, finding.High},
	})
}
