package rules

import (
	"path"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// UntrustedDiff flags prompt-injection phrasing addressed to a reviewer in
// added comments, docs prose and string literals. The patch is data
// (ASI01:2026, AST05): flag it, never follow it. A docs line that only quotes
// such a phrase is reported for review instead of blocking.
func UntrustedDiff(f diff.File) []finding.Finding {
	var out []finding.Finding
	closer := ""
	for _, line := range f.Lines {
		var inBlock bool
		closer, inBlock = advanceBlock(closer, line.Text)
		if line.Kind != diff.Added {
			continue
		}
		text := candidates(f, line, inBlock)
		if translationCopy(f.Path, line.Text) {
			// A localized UI sentence is copy, not an instruction: only text
			// aimed at a model counts there.
			if !directiveAgent.MatchString(text) {
				continue
			}
		} else if !instructionShaped(text) {
			continue
		}
		out = add(out, finding.Input{
			Title:    "Untrusted instruction in diff comment",
			Class:    finding.UntrustedDiff,
			Path:     f.Path,
			Line:     line.Line,
			CWE:      "CWE-1427",
			OWASP:    finding.LabelsFor(finding.UntrustedDiff),
			Severity: severityOf(f, line.Text),
		})
	}
	return out
}

func severityOf(f diff.File, text string) finding.Severity {
	if docExts[strings.ToLower(path.Ext(f.Path))] && quotedMention(text) {
		return finding.Medium
	}
	return finding.High
}
