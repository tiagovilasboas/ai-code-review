package review

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/rules"
)

const sample = "diff --git a/b.ts b/b.ts\n+++ b/b.ts\n@@ -1 +1,2 @@\n+x\n+y\n" +
	"diff --git a/a.ts b/a.ts\n+++ b/a.ts\n@@ -1 +1 @@\n+z\n"

func mk(sev finding.Severity, path string, line int, cwe string) finding.Finding {
	f, ok := finding.New(finding.Input{
		Title: "t", Class: finding.XSS, Path: path, Line: line, CWE: cwe,
		OWASP: finding.LabelsFor(finding.XSS), Severity: sev,
	})
	if !ok {
		panic("invalid fixture finding")
	}
	return f
}

func TestSortAndDedupe(t *testing.T) {
	rule := func(f diff.File) []finding.Finding {
		var out []finding.Finding
		for _, l := range f.Added {
			sev := finding.Low
			if l.Text == "z" {
				sev = finding.Critical
			}
			out = append(out, mk(sev, l.Path, l.Line, "CWE-79"))
		}
		return out
	}
	dup := func(f diff.File) []finding.Finding {
		// Same path:line:cwe as rule but a different severity; first wins.
		return []finding.Finding{mk(finding.Critical, f.Path, 1, "CWE-79")}
	}
	got := Run(sample, []rules.Rule{rule, dup})
	want := []struct {
		path string
		line int
		sev  finding.Severity
	}{
		{"a.ts", 1, finding.Critical},
		{"b.ts", 1, finding.Low},
		{"b.ts", 2, finding.Low},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i, w := range want {
		if got[i].Path != w.path || got[i].Line != w.line || got[i].Severity != w.sev {
			t.Errorf("[%d] got %+v, want %+v", i, got[i], w)
		}
	}
}

func TestDifferentCWEKept(t *testing.T) {
	rule := func(f diff.File) []finding.Finding {
		return []finding.Finding{mk(finding.High, f.Path, 1, "CWE-79"), mk(finding.High, f.Path, 1, "CWE-918")}
	}
	if got := Run(sample, []rules.Rule{rule}); len(got) != 4 {
		t.Fatalf("got %d findings, want 4", len(got))
	}
}

func TestNoRulesNoFindings(t *testing.T) {
	if got := Run(sample, nil); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
