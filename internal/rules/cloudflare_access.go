package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	accessEmailHeader = regexp.MustCompile(`(?i)cf-access-authenticated-user-email`)
	accessJWTHeader   = regexp.MustCompile(`(?i)cf-access-jwt-assertion`)
	workerCodeExt     = regexp.MustCompile(`(?i)\.(?:[cm]?[jt]sx?|py|go|php|rb)$`)
)

// accessHeaderFindings flags Worker or origin code that reads the Access
// identity header without validating the Access JWT in the same file's diff.
// Validation may live elsewhere, so this is HIGH but REVIEW, never BLOCK.
func accessHeaderFindings(f diff.File) []finding.Finding {
	if !workerCodeExt.MatchString(f.Path) {
		return nil
	}
	for _, l := range f.Lines {
		if accessJWTHeader.MatchString(l.Text) {
			return nil
		}
	}
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) || !accessEmailHeader.MatchString(l.Text) {
			continue
		}
		in := cfConfig(f, l.Line, "Cloudflare Access identity header trusted without JWT validation", "CWE-290", finding.High, "A01:2021")
		if fnd, ok := finding.New(in); ok {
			fnd.Action = finding.Review
			out = append(out, fnd)
		}
	}
	return out
}
