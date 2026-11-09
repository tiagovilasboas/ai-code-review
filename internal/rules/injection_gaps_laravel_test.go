package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestLaravelRawSQL(t *testing.T) {
	const ctl = "app/Http/Controllers/UserController.php"
	runRuleCases(t, SQLInjection, finding.SQLInjection, "CWE-89", []ruleCase{
		{"DB::select interpolation", mkFile(ctl,
			`+    $rows = DB::select("SELECT * FROM users WHERE name = '{$request->input('name')}'");`,
		), []int{1}, finding.High},
		{"DB::select concatenation", mkFile(ctl,
			`+    $rows = DB::select("SELECT * FROM users WHERE id = " . $request->get('id'));`,
		), []int{1}, finding.High},
		{"whereRaw interpolation", mkFile(ctl,
			`+    $q = User::whereRaw("price > {$request->input('min')}")->get();`,
		), []int{1}, finding.High},
		{"DB::raw interpolation", mkFile(ctl,
			`+    $q = User::select(DB::raw("count(*) as {$request->input('alias')}"))->get();`,
		), []int{1}, finding.High},
		{"orderByRaw non request", mkFile("app/Models/Report.php",
			`+    return $query->orderByRaw("FIELD(status, $statuses)");`,
		), []int{1}, finding.Medium},
		{"DB::select bindings", mkFile(ctl,
			`+    $rows = DB::select('SELECT * FROM users WHERE id = ?', [$request->input('id')]);`,
		), nil, finding.Medium},
		{"whereRaw bindings", mkFile(ctl,
			`+    $q = User::whereRaw('price > ?', [$request->input('min')])->get();`,
		), nil, finding.Medium},
		{"DB::select named bindings", mkFile(ctl,
			`+    $rows = DB::select("SELECT * FROM users WHERE id = :id", ['id' => $id]);`,
		), nil, finding.Medium},
	})
}
