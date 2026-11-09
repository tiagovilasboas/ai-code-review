package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// mkFile builds a diff.File from hunk lines prefixed "+" (added) or " "
// (context). New-file line numbers start at 1 and increment per line.
func mkFile(path string, lines ...string) diff.File {
	f := diff.File{Path: path}
	for i, raw := range lines {
		l := diff.Line{Path: path, Line: i + 1, Text: raw[1:], Kind: diff.Context}
		if raw[0] == '+' {
			l.Kind = diff.Added
			f.Added = append(f.Added, l)
		}
		f.Lines = append(f.Lines, l)
	}
	return f
}

// wantFinding asserts one finding's identifying fields.
func wantFinding(t *testing.T, got finding.Finding, title string, class finding.Class, line int, cwe string, sev finding.Severity, act finding.Action) {
	t.Helper()
	if got.Title != title || got.Class != class || got.Line != line || got.CWE != cwe || got.Severity != sev || got.Action != act {
		t.Fatalf("unexpected finding: %+v", got)
	}
	want := finding.LabelsFor(class)
	if len(got.OWASP) != len(want) {
		t.Fatalf("owasp = %v, want %v", got.OWASP, want)
	}
	for i := range want {
		if got.OWASP[i] != want[i] {
			t.Fatalf("owasp = %v, want %v", got.OWASP, want)
		}
	}
}
