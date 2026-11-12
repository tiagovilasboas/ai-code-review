// Package report renders findings as the plain-text review report.
package report

import (
	"fmt"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/triage"
)

const noFindings = "no evidence-based findings\n"

// PickMergeBlocker prefers AuthZ/IDOR, then secrets, then the first BLOCK.
// It returns false when there is no BLOCK finding.
func PickMergeBlocker(findings []finding.Finding) (finding.Finding, bool) {
	var first, secret *finding.Finding
	for i := range findings {
		f := &findings[i]
		if f.Action != finding.Block {
			continue
		}
		if f.Class == finding.AuthZ {
			return *f, true
		}
		if f.Class == finding.Secret && secret == nil {
			secret = f
		}
		if first == nil {
			first = f
		}
	}
	switch {
	case secret != nil:
		return *secret, true
	case first != nil:
		return *first, true
	}
	return finding.Finding{}, false
}

// MergeDecision returns the two-line decision header, or "" without a blocker.
func MergeDecision(findings []finding.Finding) string {
	b, ok := PickMergeBlocker(findings)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Decision: DO NOT MERGE\nReason: %s at %s:%d (%s)",
		b.Class.Label(), b.Path, b.Line, asvsOrOfficial(b))
}

func asvsOrOfficial(f finding.Finding) string {
	for _, id := range f.OWASP {
		if strings.HasPrefix(id, "ASVS-5.0-") {
			return id
		}
	}
	if len(f.OWASP) > 0 {
		return f.OWASP[0]
	}
	return ""
}

func formatFinding(f finding.Finding) string {
	lines := []string{
		"Finding: " + f.Title,
		"Class: " + f.Class.Label(),
		fmt.Sprintf("Evidence: %s:%d", f.Path, f.Line),
		"CWE: " + f.CWE,
		"OWASP: " + strings.Join(f.OWASP, ", "),
		"Severity: " + string(f.Severity),
		"Action: " + string(f.Action),
	}
	if f.Action == finding.Block && f.Class.ChangesMerge() {
		lines = append(lines, "Decision: DO NOT MERGE")
	}
	return strings.Join(lines, "\n")
}

// FormatText renders the text report; internal/cli/testdata/golden pins its output.
func FormatText(findings []finding.Finding) string {
	return FormatTextTriaged(findings, nil)
}

// FormatTextTriaged renders the text report followed by the Triaged section.
func FormatTextTriaged(findings []finding.Finding, applied []triage.Applied) string {
	out := noFindings
	if len(findings) > 0 {
		blocks := make([]string, len(findings))
		for i, f := range findings {
			blocks[i] = formatFinding(f)
		}
		out = MergeDecision(findings) + "\n\n" + strings.Join(blocks, "\n\n") + "\n"
	}
	return out + formatTriaged(applied)
}

// ExitCode is 1 when any BLOCK finding exists, else 0.
func ExitCode(findings []finding.Finding) int {
	for _, f := range findings {
		if f.Action == finding.Block {
			return 1
		}
	}
	return 0
}
