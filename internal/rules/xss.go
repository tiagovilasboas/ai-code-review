package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// xssTestPath matches test, mock and story files, where HTML sinks mount fixtures.
var xssTestPath = regexp.MustCompile(`(?i)(?:^|/)(?:__tests__|__mocks__|tests?|specs?|e2e|cypress|fixtures?|stories)/|\.(?:test|spec|stories)\.[cm]?[jt]sx?$`)

// XSS flags HTML sinks and unvalidated URL attributes on added, non-comment
// lines unless the value is a constant or visibly sanitized.
func XSS(f diff.File) []finding.Finding {
	if xssTestPath.MatchString(f.Path) {
		return nil
	}
	ctx := newXSSContext(f)
	backend := isBackendSource(f.Path)
	var out []finding.Finding
	serverHTML := map[int]bool{}
	for _, n := range serverHTMLHits(f) {
		serverHTML[n] = true
	}
	for _, line := range f.Added {
		if IsCommentLine(line.Text) {
			continue
		}
		domSink := ctx.flags(line.Text) && !(backend && !ctx.flags(scanExpr(line.Text, false).code))
		if !domSink && !serverHTML[line.Line] && !templateBypassHit(f.Path, line.Text) {
			continue
		}
		out = add(out, xssInput(f.Path, line.Line))
	}
	return out
}

func xssInput(path string, line int) finding.Input {
	return finding.Input{
		Title:    "DOM XSS",
		Class:    finding.XSS,
		Path:     path,
		Line:     line,
		CWE:      "CWE-79",
		OWASP:    finding.LabelsFor(finding.XSS),
		Severity: finding.High,
	}
}

// isBackendSource reports server-side languages where a DOM sink name can only
// appear as text (a rule pattern, a log message), never as a browser API call.
// There the sink must occur in code, outside string literals.
func isBackendSource(p string) bool {
	switch langOf(p) {
	case langGo, langPython, langJava:
		return true
	}
	return false
}
