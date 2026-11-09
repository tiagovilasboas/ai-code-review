package bench

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func k(path string, line int, class string) Key { return Key{Path: path, Line: line, Class: class} }

func TestMatchRequiresPathLineAndClass(t *testing.T) {
	expected := []Key{k("a.js", 3, "xss"), k("b.js", 1, "ssrf"), k("c.js", 9, "authz")}
	got := []Key{k("a.js", 3, "xss"), k("b.js", 2, "ssrf"), k("c.js", 9, "secret"), k("a.js", 3, "xss")}
	hit, missed, extra := Match(expected, got)
	if len(hit) != 1 || hit[0] != k("a.js", 3, "xss") {
		t.Fatalf("hit = %v", hit)
	}
	if len(missed) != 2 || len(extra) != 2 {
		t.Fatalf("missed = %v, extra = %v", missed, extra)
	}
}

func TestMatchDuplicatesCountOnce(t *testing.T) {
	hit, missed, extra := Match([]Key{k("a", 1, "xss"), k("a", 1, "xss")}, nil)
	if len(hit) != 0 || len(missed) != 1 || len(extra) != 0 {
		t.Fatalf("hit=%v missed=%v extra=%v", hit, missed, extra)
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCountsMetrics(t *testing.T) {
	c := Counts{TP: 6, FP: 2, FN: 4}
	p, _ := c.Precision()
	r, _ := c.Recall()
	f, ok := c.F1()
	if !near(p, 0.75) || !near(r, 0.6) || !ok || !near(f, 2*0.75*0.6/1.35) {
		t.Fatalf("p=%v r=%v f=%v", p, r, f)
	}
}

func TestCountsUndefined(t *testing.T) {
	if _, ok := (Counts{FN: 3}).Precision(); ok {
		t.Error("precision with no reports should be undefined")
	}
	if _, ok := (Counts{FP: 3}).Recall(); ok {
		t.Error("recall with nothing expected should be undefined")
	}
	if _, ok := (Counts{FP: 1}).F1(); ok {
		t.Error("f1 should be undefined")
	}
	if f, ok := (Counts{FP: 1, FN: 1}).F1(); !ok || f != 0 {
		t.Errorf("f1 = %v %v, want 0 true", f, ok)
	}
}

func TestEvaluateTalliesPerClass(t *testing.T) {
	cases := []Case{
		{ID: "tp", Diff: "d1", Expected: []Key{k("a", 1, "xss")}},
		{ID: "fn", Diff: "d2", Expected: []Key{k("b", 1, "ssrf")}},
		{ID: "fp", Diff: "d3"},
		{ID: "quiet", Diff: "d4"},
	}
	out := map[string][]Key{
		"d1": {k("a", 1, "xss")},
		"d3": {k("c", 2, "authz")},
	}
	r := Evaluate(cases, func(d string) []Key { return out[d] })
	if r.Overall != (Counts{TP: 1, FP: 1, FN: 1}) {
		t.Fatalf("overall = %+v", r.Overall)
	}
	if r.CleanCases != 2 || r.CleanSilent != 1 {
		t.Fatalf("clean = %d silent = %d", r.CleanCases, r.CleanSilent)
	}
	want := map[string]Counts{"xss": {TP: 1}, "ssrf": {FN: 1}, "authz": {FP: 1}}
	for _, c := range r.Classes {
		if w, ok := want[c.Class]; ok && c.Counts != w {
			t.Errorf("%s = %+v, want %+v", c.Class, c.Counts, w)
		}
	}
	if !strings.Contains(Markdown(r), "| **overall** | 1 | 1 | 1 | 0.50 | 0.50 | 0.50 |") {
		t.Errorf("markdown missing overall row:\n%s", Markdown(r))
	}
}

func writeCase(t *testing.T, root, id, expected string) {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"change.diff": "diff\n", "expected.json": expected} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadCorpus(t *testing.T) {
	root := t.TempDir()
	writeCase(t, root, "b-case", `{"description":"d","source":"s","findings":[]}`)
	writeCase(t, root, "a-case", `{"description":"d","source":"s","findings":[{"path":"x.js","line":2,"class":"xss"}]}`)
	cases, err := LoadCorpus(root)
	if err != nil || len(cases) != 2 || cases[0].ID != "a-case" || len(cases[0].Expected) != 1 {
		t.Fatalf("cases = %+v err = %v", cases, err)
	}
}

func TestLoadCorpusErrors(t *testing.T) {
	bad := map[string]string{
		"unknown-class": `{"findings":[{"path":"x","line":1,"class":"nope"}]}`,
		"zero-line":     `{"findings":[{"path":"x","line":0,"class":"xss"}]}`,
		"bad-json":      `{`,
	}
	for id, body := range bad {
		root := t.TempDir()
		writeCase(t, root, id, body)
		if _, err := LoadCorpus(root); err == nil {
			t.Errorf("%s: expected error", id)
		}
	}
	if _, err := LoadCorpus(t.TempDir()); err == nil {
		t.Error("empty corpus should error")
	}
	if _, err := LoadCorpus(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("missing corpus should error")
	}
}
