package report

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func mk(t *testing.T, c finding.Class, sev finding.Severity, path string, line int) finding.Finding {
	t.Helper()
	f, ok := finding.New(finding.Input{
		Title: "T-" + string(c), Class: c, Path: path, Line: line, CWE: "CWE-1",
		OWASP: finding.LabelsFor(c), Severity: sev,
	})
	if !ok {
		t.Fatal("invalid fixture")
	}
	return f
}

func TestEmpty(t *testing.T) {
	if got := FormatText(nil); got != "no evidence-based findings\n" {
		t.Fatalf("got %q", got)
	}
	if MergeDecision(nil) != "" || ExitCode(nil) != 0 {
		t.Fatal("empty must be clean")
	}
}

func TestPickMergeBlockerPriority(t *testing.T) {
	x := mk(t, finding.XSS, finding.High, "a", 1)
	s := mk(t, finding.Secret, finding.High, "b", 2)
	a := mk(t, finding.AuthZ, finding.High, "c", 3)
	low := mk(t, finding.AuthZ, finding.Low, "d", 4) // REVIEW, never a blocker

	got, ok := PickMergeBlocker([]finding.Finding{low, x, s, a})
	if !ok || got.Class != finding.AuthZ || got.Path != "c" {
		t.Fatalf("authz first: %+v", got)
	}
	got, _ = PickMergeBlocker([]finding.Finding{x, s})
	if got.Class != finding.Secret {
		t.Fatalf("secret second: %+v", got)
	}
	got, _ = PickMergeBlocker([]finding.Finding{low, x})
	if got.Class != finding.XSS {
		t.Fatalf("first block: %+v", got)
	}
	if _, ok := PickMergeBlocker([]finding.Finding{low}); ok {
		t.Fatal("REVIEW only has no blocker")
	}
}

func TestFormatTextExact(t *testing.T) {
	a := mk(t, finding.AuthZ, finding.High, "src/o.ts", 36)
	x := mk(t, finding.XSS, finding.Medium, "w.tsx", 18)
	want := "Decision: DO NOT MERGE\n" +
		"Reason: AuthZ/IDOR at src/o.ts:36 (ASVS-5.0-8.2.2)\n\n" +
		"Finding: T-authz\nClass: AuthZ/IDOR\nEvidence: src/o.ts:36\nCWE: CWE-1\n" +
		"OWASP: A01:2021, API1:2023, ASVS-5.0-8.2.2\nSeverity: HIGH\nAction: BLOCK\nDecision: DO NOT MERGE\n\n" +
		"Finding: T-xss\nClass: XSS / HTML sink\nEvidence: w.tsx:18\nCWE: CWE-1\n" +
		"OWASP: A03:2021, ASVS-5.0-1.3.1\nSeverity: MEDIUM\nAction: REVIEW\n"
	if got := FormatText([]finding.Finding{a, x}); got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestNonMergeChangingBlockOmitsPerFindingDecision(t *testing.T) {
	x := mk(t, finding.XSS, finding.High, "w.tsx", 1)
	got := FormatText([]finding.Finding{x})
	want := "Decision: DO NOT MERGE\nReason: XSS / HTML sink at w.tsx:1 (ASVS-5.0-1.3.1)\n\n" +
		"Finding: T-xss\nClass: XSS / HTML sink\nEvidence: w.tsx:1\nCWE: CWE-1\n" +
		"OWASP: A03:2021, ASVS-5.0-1.3.1\nSeverity: HIGH\nAction: BLOCK\n"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestExitCode(t *testing.T) {
	rev := mk(t, finding.XSS, finding.Low, "a", 1)
	blk := mk(t, finding.XSS, finding.Critical, "a", 2)
	if ExitCode([]finding.Finding{rev}) != 0 || ExitCode([]finding.Finding{rev, blk}) != 1 {
		t.Fatal("exit code")
	}
}
