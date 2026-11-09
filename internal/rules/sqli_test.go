package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// ruleCase is one file shown to a rule and the findings it must produce.
type ruleCase struct {
	name string
	file diff.File
	want []int // line numbers, in order
	sev  finding.Severity
}

func runRuleCases(t *testing.T, rule Rule, class finding.Class, cwe string, cases []ruleCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rule(tc.file)
			if len(got) != len(tc.want) {
				t.Fatalf("want %d findings at %v, got %+v", len(tc.want), tc.want, got)
			}
			for i, line := range tc.want {
				act := finding.Review
				if tc.sev == finding.High {
					act = finding.Block
				}
				wantFinding(t, got[i], got[i].Title, class, line, cwe, tc.sev, act)
			}
		})
	}
}

func TestSQLInjectionFires(t *testing.T) {
	runRuleCases(t, SQLInjection, finding.SQLInjection, "CWE-89", []ruleCase{
		{"js template literal from req", mkFile("src/api/users.ts",
			` export async function find(req, res) {`,
			`+  const rows = await db.query(`+"`SELECT * FROM users WHERE name = '${req.query.name}'`"+`);`,
		), []int{2}, finding.High},
		{"js concatenation via tainted var", mkFile("src/api/users.js",
			`+  const id = req.params.id;`,
			`+  const rows = await pool.query("SELECT * FROM orders WHERE id = " + id);`,
		), []int{2}, finding.High},
		{"js built sql var, not request", mkFile("src/db/report.ts",
			`+  const sql = "SELECT * FROM t ORDER BY " + column;`,
			`+  return knex.raw(sql);`,
		), []int{2}, finding.Medium},
		{"js multi-line call", mkFile("src/api/a.ts",
			`+  const { name } = req.body;`,
			`+  await sequelize.query(`,
			`+    `+"`DELETE FROM users WHERE name = '${name}'`",
			`+  );`,
		), []int{2}, finding.High},
		{"python f-string from request.args", mkFile("app/views.py",
			`+    uid = request.args.get("id")`,
			`+    cur.execute(f"SELECT * FROM users WHERE id = {uid}")`,
		), []int{2}, finding.High},
		{"python percent format", mkFile("app/db.py",
			`+    cursor.execute("SELECT * FROM t WHERE name = '%s'" % name)`,
		), []int{1}, finding.Medium},
		{"python str.format", mkFile("app/db.py",
			`+    cursor.execute("DELETE FROM t WHERE id = {}".format(item_id))`,
		), []int{1}, finding.Medium},
		{"python appended sql var", mkFile("app/db.py",
			`+    q = "SELECT * FROM t WHERE 1 = 1"`,
			`+    q += " AND owner = '" + request.form["owner"] + "'"`,
			`+    cursor.execute(q)`,
		), []int{3}, finding.High},
		{"go Sprintf from query string", mkFile("internal/users/handler.go",
			`+	name := r.URL.Query().Get("name")`,
			`+	rows, err := db.Query(fmt.Sprintf("SELECT id FROM users WHERE name = '%s'", name))`,
		), []int{2}, finding.High},
		{"go concatenation", mkFile("internal/store/store.go",
			`+	_, err := db.Exec("UPDATE t SET flag = 1 WHERE id = " + id)`,
		), []int{1}, finding.Medium},
		{"java spring param", mkFile("src/main/java/UserController.java",
			`+    public List<User> find(@RequestParam String name) {`,
			`+        return em.createQuery("SELECT u FROM User u WHERE u.name = '" + name + "'").getResultList();`,
		), []int{2}, finding.High},
		{"java statement executeQuery", mkFile("src/main/java/Repo.java",
			`+        String sql = "SELECT * FROM t WHERE id = " + id;`,
			`+        ResultSet rs = stmt.executeQuery(sql);`,
		), []int{2}, finding.Medium},
		{"php interpolation from $_GET", mkFile("public/user.php",
			`+$id = $_GET['id'];`,
			`+$result = mysqli_query($conn, "SELECT * FROM users WHERE id = $id");`,
		), []int{2}, finding.High},
		{"php concatenation", mkFile("public/user.php",
			`+$rows = $pdo->query("SELECT * FROM t WHERE name = '" . $name . "'");`,
		), []int{1}, finding.Medium},
	})
}

func TestSQLInjectionTrueNegatives(t *testing.T) {
	runRuleCases(t, SQLInjection, finding.SQLInjection, "CWE-89", []ruleCase{
		{"js parameterized", mkFile("src/api/users.ts",
			`+  const id = req.params.id;`,
			`+  const rows = await db.query("SELECT * FROM users WHERE id = $1", [id]);`,
		), nil, ""},
		{"js literal interpolation", mkFile("src/api/users.ts",
			`+  await db.query(`+"`SELECT * FROM ${'users'} WHERE active = ${1}`"+`);`,
		), nil, ""},
		{"js constant table", mkFile("src/db/x.ts",
			`+  await db.query("SELECT * FROM " + USERS_TABLE + " WHERE id = ?", [id]);`,
		), nil, ""},
		{"js escaped value", mkFile("src/db/x.ts",
			`+  await db.query("SELECT * FROM t WHERE id = " + mysql.escape(req.query.id));`,
		), nil, ""},
		{"js comment", mkFile("src/db/x.ts",
			`+  // db.query("SELECT * FROM t WHERE id = " + req.query.id)`,
		), nil, ""},
		{"js dynamic non-sql string", mkFile("src/db/x.ts",
			`+  re.exec("hello " + name);`,
		), nil, ""},
		{"python parameterized", mkFile("app/db.py",
			`+    uid = request.args.get("id")`,
			`+    cur.execute("SELECT * FROM users WHERE id = %s", (uid,))`,
		), nil, ""},
		{"python sql var reassigned to literal", mkFile("app/db.py",
			`+    q = "SELECT * FROM t WHERE id = " + uid`,
			`+    q = "SELECT * FROM t WHERE id = %s"`,
			`+    cursor.execute(q, (uid,))`,
		), nil, ""},
		{"go parameterized", mkFile("internal/users/handler.go",
			`+	name := r.URL.Query().Get("name")`,
			`+	rows, err := db.Query("SELECT id FROM users WHERE name = $1", name)`,
		), nil, ""},
		{"java prepared statement", mkFile("src/main/java/Repo.java",
			`+        PreparedStatement ps = conn.prepareStatement("SELECT * FROM t WHERE id = ?");`,
			`+        ps.setString(1, request.getParameter("id"));`,
		), nil, ""},
		{"php prepared", mkFile("public/user.php",
			`+$stmt = $pdo->prepare('SELECT * FROM users WHERE id = ?');`,
			`+$stmt->execute([$_GET['id']]);`,
		), nil, ""},
		{"php single quotes do not interpolate", mkFile("public/user.php",
			`+$rows = $pdo->query('SELECT * FROM t WHERE name = $name');`,
		), nil, ""},
		{"unknown language", mkFile("db/query.sql",
			`+SELECT * FROM t WHERE id = " + id`,
		), nil, ""},
	})
}
