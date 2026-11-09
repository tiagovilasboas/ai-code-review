package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// instructionShaped matches prompt-injection phrasing. The patch is data
// (ASI01:2026, AST05): flag it, never follow it.
var instructionShaped = regexp.MustCompile(`(?i)(?:ignore\s+(?:all\s+)?(?:previous|prior|above)\s+instructions|you\s+must\s+merge|merge\s+this\s+(?:pr|pull request)|approve\s+this\s+(?:pr|pull request)|skip\s+(?:this\s+)?review|do\s+not\s+report|disregard\s+(?:the\s+)?(?:guardrail|evidence))`)

// UntrustedDiff flags instruction-like text inside added comment lines.
func UntrustedDiff(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, line := range f.Added {
		if !IsCommentLine(line.Text) || !instructionShaped.MatchString(line.Text) {
			continue
		}
		out = add(out, finding.Input{
			Title:    "Untrusted instruction in diff comment",
			Class:    finding.UntrustedDiff,
			Path:     f.Path,
			Line:     line.Line,
			CWE:      "CWE-1427",
			OWASP:    finding.LabelsFor(finding.UntrustedDiff),
			Severity: finding.High,
		})
	}
	return out
}
