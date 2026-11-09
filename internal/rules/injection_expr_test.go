package rules

import (
	"reflect"
	"testing"
)

func TestScanExprInterpolations(t *testing.T) {
	cases := []struct {
		name, in string
		php      bool
		interps  []string
		code     string
	}{
		{"js template", "`a ${x} b ${'lit'}`", false, []string{"x", "'lit'"}, `""`},
		{"python f-string", `f"id = {uid} {{raw}}"`, false, []string{"uid"}, `""`},
		{"python plain string keeps braces", `"id = {uid}"`, false, nil, `""`},
		{"php double quotes", `"id = $id and {$row['x']}"`, true, []string{"$row['x']", "$id"}, `""`},
		{"php single quotes", `'id = $id'`, true, nil, `""`},
		{"concatenation", `"a" + b + 'c'`, false, nil, `"" + b + ""`},
		{"python triple quotes", `"""SELECT "x" FROM t""" % y`, false, nil, `"" % y`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := scanExpr(tc.in, tc.php)
			if !sameStrings(e.interps, tc.interps) || e.code != tc.code {
				t.Fatalf("scanExpr(%q) = interps %q code %q", tc.in, e.interps, e.code)
			}
		})
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

func TestCallArgs(t *testing.T) {
	text := `query("a, b", [x, y], { k: f(1, 2) }) + rest(z)`
	got := callArgs(text, len("query"))
	want := []string{`"a, b"`, `[x, y]`, `{ k: f(1, 2) }`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("callArgs = %q, want %q", got, want)
	}
}

func TestOperandsSkipConstantsAndBuilders(t *testing.T) {
	got := operands(`fmt.Sprintf("", id, MAX_ROWS, 10, nil) + user.name`)
	want := []string{"id", "user.name"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("operands = %q, want %q", got, want)
	}
}

func TestMentionsWholeWordOnly(t *testing.T) {
	for code, want := range map[string]bool{
		"q(id)":      true,
		"q(user.id)": false,
		"q(ids)":     false,
		"q($id)":     false,
		"q(x, id)":   true,
	} {
		if got := mentions(code, "id"); got != want {
			t.Errorf("mentions(%q, id) = %v, want %v", code, got, want)
		}
	}
}

func TestTaintedNamesFollowAssignments(t *testing.T) {
	f := mkFile("a.go",
		`+	name := r.URL.Query().Get("name")`,
		`+	p := filepath.Join(base, name)`,
		`+	q, err := strconv.Atoi(r.FormValue("n"))`,
		`+	other := "x"`,
	)
	got := taintedNames(f.Lines, false)
	for _, n := range []string{"name", "p", "q", "err"} {
		if !got[n] {
			t.Errorf("%s not tainted: %v", n, got)
		}
	}
	if got["other"] {
		t.Errorf("literal assignment tainted: %v", got)
	}
}
