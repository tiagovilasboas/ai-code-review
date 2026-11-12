package triage

import (
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func mk(t *testing.T, c finding.Class, path string, line int, sev finding.Severity) finding.Finding {
	t.Helper()
	f, ok := finding.New(finding.Input{
		Title: "t", Class: c, Path: path, Line: line, CWE: "CWE-79",
		OWASP: finding.LabelsFor(c), Severity: sev,
	})
	if !ok {
		t.Fatal("invalid fixture finding")
	}
	return f
}

func TestApply(t *testing.T) {
	xss := mk(t, finding.XSS, "a.ts", 3, finding.High)
	xssID := finding.ID(xss)
	other := mk(t, finding.XSS, "b.ts", 4, finding.Medium)
	sec := mk(t, finding.Secret, "c.ts", 5, finding.Critical)
	inj := mk(t, finding.UntrustedDiff, "d.md", 6, finding.High)
	all := []finding.Finding{xss, other, sec, inj}

	cases := []struct {
		name     string
		verdicts []Verdict
		wantErr  string
		wantLen  int
		check    func(t *testing.T, out []finding.Finding)
	}{
		{"drop", []Verdict{{ID: xssID, Action: Drop, Reason: "test fixture"}}, "", 3, nil},
		{"downgrade lowers severity and action", []Verdict{{ID: xssID, Action: Downgrade, Severity: finding.Medium, Reason: "r"}}, "", 4,
			func(t *testing.T, out []finding.Finding) {
				for _, f := range out {
					if f.Path == "a.ts" && (f.Severity != finding.Medium || f.Action != finding.Review) {
						t.Errorf("got %+v", f)
					}
				}
			}},
		{"raise rejected", []Verdict{{ID: finding.ID(other), Action: Downgrade, Severity: finding.High, Reason: "r"}}, "cannot raise", 0, nil},
		{"same severity rejected", []Verdict{{ID: xssID, Action: Downgrade, Severity: finding.High, Reason: "r"}}, "cannot raise", 0, nil},
		{"unknown id rejected", []Verdict{{ID: "deadbeefdeadbeef", Action: Drop, Reason: "r"}}, "cannot add", 0, nil},
		{"secret is immune to drop", []Verdict{{ID: finding.ID(sec), Action: Drop, Reason: "r"}}, "cannot be dropped", 0, nil},
		{"secret is immune to downgrade", []Verdict{{ID: finding.ID(sec), Action: Downgrade, Severity: finding.Low, Reason: "r"}}, "cannot be dropped", 0, nil},
		{"untrusted-diff is immune", []Verdict{{ID: finding.ID(inj), Action: Drop, Reason: "r"}}, "cannot be dropped", 0, nil},
		{"one bad verdict fails the batch", []Verdict{
			{ID: xssID, Action: Drop, Reason: "r"}, {ID: "nope", Action: Drop, Reason: "r"}}, "no such finding", 0, nil},
		{"invalid severity rejected", []Verdict{{ID: xssID, Action: Downgrade, Severity: "BOGUS", Reason: "r"}}, "needs a severity", 0, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, applied, err := Apply(all, File{SchemaVersion: 1, Verdicts: tc.verdicts})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) || out != nil || applied != nil {
					t.Fatalf("err = %v out=%v", err, out)
				}
				return
			}
			if err != nil || len(out) != tc.wantLen || len(applied) != len(tc.verdicts) {
				t.Fatalf("err=%v out=%d applied=%d", err, len(out), len(applied))
			}
			if tc.check != nil {
				tc.check(t, out)
			}
		})
	}
}

func TestApplyNeverAdds(t *testing.T) {
	in := []finding.Finding{mk(t, finding.XSS, "a.ts", 1, finding.Low)}
	out, applied, err := Apply(in, File{SchemaVersion: 1})
	if err != nil || len(out) != 1 || applied == nil || len(applied) != 0 {
		t.Fatalf("%v %v %v", out, applied, err)
	}
}

func TestApplyAmbiguousID(t *testing.T) {
	a := mk(t, finding.XSS, "a.ts", 1, finding.Low)
	_, _, err := Apply([]finding.Finding{a, a}, File{Verdicts: []Verdict{{ID: finding.ID(a), Action: Drop, Reason: "r"}}})
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("%v", err)
	}
}

func TestApplyResorts(t *testing.T) {
	a := mk(t, finding.XSS, "z.ts", 1, finding.High)
	b := mk(t, finding.XSS, "a.ts", 1, finding.Medium)
	out, _, err := Apply([]finding.Finding{a, b}, File{Verdicts: []Verdict{{ID: finding.ID(a), Action: Downgrade, Severity: finding.Medium, Reason: "r"}}})
	if err != nil || out[0].Path != "a.ts" {
		t.Fatalf("%v %v", out, err)
	}
}

func TestParse(t *testing.T) {
	ok := `{"schemaVersion":1,"verdicts":[{"id":"x","action":"drop","reason":"r"}]}`
	bad := map[string]string{
		"not json":              `nope`,
		"empty":                 ``,
		"wrong version":         `{"schemaVersion":2,"verdicts":[]}`,
		"missing version":       `{"verdicts":[]}`,
		"unknown field":         `{"schemaVersion":1,"verdicts":[],"findings":[{"id":"new"}]}`,
		"unknown verdict field": `{"schemaVersion":1,"verdicts":[{"id":"x","action":"drop","reason":"r","title":"new"}]}`,
		"trailing data":         ok + `{}`,
		"no reason":             `{"schemaVersion":1,"verdicts":[{"id":"x","action":"drop"}]}`,
		"no id":                 `{"schemaVersion":1,"verdicts":[{"action":"drop","reason":"r"}]}`,
		"unknown action":        `{"schemaVersion":1,"verdicts":[{"id":"x","action":"raise","reason":"r"}]}`,
		"drop w/ severity":      `{"schemaVersion":1,"verdicts":[{"id":"x","action":"drop","severity":"LOW","reason":"r"}]}`,
		"downgrade w/o sev":     `{"schemaVersion":1,"verdicts":[{"id":"x","action":"downgrade","reason":"r"}]}`,
		"duplicate id":          `{"schemaVersion":1,"verdicts":[{"id":"x","action":"drop","reason":"r"},{"id":"x","action":"drop","reason":"r"}]}`,
		"long reason":           `{"schemaVersion":1,"verdicts":[{"id":"x","action":"drop","reason":"` + strings.Repeat("a", 501) + `"}]}`,
	}
	if f, err := Parse([]byte(ok)); err != nil || len(f.Verdicts) != 1 {
		t.Fatalf("%v %v", f, err)
	}
	for name, in := range bad {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
