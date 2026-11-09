package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// xssSinks are HTML-injecting APIs; every hit shares the title "DOM XSS".
var xssSinks = []*regexp.Regexp{
	regexp.MustCompile(`dangerouslySetInnerHTML`),
	regexp.MustCompile(`\.innerHTML\s*=`),
	regexp.MustCompile(`\.outerHTML\s*=`),
	regexp.MustCompile(`\binsertAdjacentHTML\s*\(`),
	regexp.MustCompile(`\bdocument\.write\s*\(`),
	regexp.MustCompile(`\bv-html\b`),
	regexp.MustCompile(`\bbypassSecurityTrust(?:Html|Script|Url|ResourceUrl|Style)\b`),
	regexp.MustCompile(`\bunsafeHTML\s*\(`),
}

var xssSanitized = regexp.MustCompile(`__html\s*:\s*(?:DOMPurify|sanitizeHtml|sanitize)\s*\(`)

// XSS flags HTML sinks on added, non-comment lines unless sanitized on the same line.
func XSS(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, line := range f.Added {
		if IsCommentLine(line.Text) || xssSanitized.MatchString(line.Text) {
			continue
		}
		for _, re := range xssSinks {
			if re.MatchString(line.Text) {
				out = add(out, finding.Input{
					Title:    "DOM XSS",
					Class:    finding.XSS,
					Path:     f.Path,
					Line:     line.Line,
					CWE:      "CWE-79",
					OWASP:    finding.LabelsFor(finding.XSS),
					Severity: finding.High,
				})
				break
			}
		}
	}
	return out
}
