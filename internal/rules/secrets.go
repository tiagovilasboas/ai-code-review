package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	secretAssign = regexp.MustCompile(`(?i)(?:password|passwd|secret|token|api[_-]?key|private[_-]?key|access[_-]?key|credential)\s*[:=]\s*["'][^"']+["']`)
	knownPrefix  = regexp.MustCompile(`["'](?:pk_live_|pk_test_|sk_live_|sk_test_|AKIA|ghp_|github_pat_|xox[baprs]-)`)
	corsStar     = regexp.MustCompile(`\bcors\s*\(\s*\{[^}]*origin\s*:\s*["']\*["']`)
)

// Secrets flags hardcoded credentials (CWE-798) and wildcard CORS origins (CWE-942)
// on added lines. Lines reading process.env and comments are ignored.
func Secrets(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, line := range f.Added {
		if IsCommentLine(line.Text) || strings.Contains(line.Text, "process.env") {
			continue
		}
		if secretAssign.MatchString(line.Text) || knownPrefix.MatchString(line.Text) {
			out = add(out, finding.Input{
				Title:    "Hardcoded credential in source",
				Class:    finding.Secret,
				Path:     f.Path,
				Line:     line.Line,
				CWE:      "CWE-798",
				OWASP:    finding.LabelsFor(finding.Secret),
				Severity: finding.Critical,
			})
		}
		if corsStar.MatchString(line.Text) {
			out = add(out, finding.Input{
				Title:    "Overly permissive CORS origin",
				Class:    finding.CORS,
				Path:     f.Path,
				Line:     line.Line,
				CWE:      "CWE-942",
				OWASP:    finding.LabelsFor(finding.CORS),
				Severity: finding.Medium,
			})
		}
	}
	return out
}
