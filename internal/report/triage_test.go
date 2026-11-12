package report

import (
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/triage"
)

func TestFormatTextTriagedExact(t *testing.T) {
	f, _ := finding.New(finding.Input{
		Title: "t", Class: finding.XSS, Path: "a.ts", Line: 3, CWE: "CWE-79",
		OWASP: finding.LabelsFor(finding.XSS), Severity: finding.High,
	})
	low := f
	low.Severity, low.Action = finding.Low, finding.Review
	id := finding.ID(f)
	got := FormatTextTriaged(nil, []triage.Applied{
		{ID: id, Verb: triage.Drop, Reason: "fixture", Before: f, After: f},
		{ID: id, Verb: triage.Downgrade, Reason: "doc only", Before: f, After: low},
	})
	want := "no evidence-based findings\n\n" + triageNotice + "\n" +
		"- drop XSS / HTML sink at a.ts:3 (" + id + ", HIGH BLOCK)\n  Reason: fixture\n" +
		"- downgrade XSS / HTML sink at a.ts:3 (" + id + ", HIGH BLOCK) -> LOW REVIEW\n  Reason: doc only\n"
	if got != want {
		t.Fatalf("got\n%q\nwant\n%q", got, want)
	}
	if strings.Contains(FormatText(nil), "Triaged") {
		t.Fatal("plain report must not mention triage")
	}
}
