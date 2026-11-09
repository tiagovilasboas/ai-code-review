package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	ssrfServerPath   = regexp.MustCompile(`(?:^|/)(?:api|server|backend|webhook|worker|jobs)/`)
	ssrfClientExt    = regexp.MustCompile(`\.(?:tsx|jsx|vue)$`)
	ssrfAPIServer    = regexp.MustCompile(`(?:^|/)(?:api|server)/`)
	ssrfServerFrames = regexp.MustCompile(`\b(?:express|fastify|koa|IncomingMessage)\b`)
	ssrfReqInput     = regexp.MustCompile(`\breq\.(?:body|query|params)\b`)
	ssrfResponseText = regexp.MustCompile(`\bresponse\.text\s*\(`)
	ssrfReturnsHTML  = regexp.MustCompile(`res\.(?:json|send)\s*\(.*\bhtml\b`)
)

func looksServerSide(f diff.File) bool {
	if ssrfServerPath.MatchString(f.Path) {
		return true
	}
	if ssrfClientExt.MatchString(f.Path) && !ssrfAPIServer.MatchString(f.Path) {
		return false
	}
	texts := make([]string, len(f.Lines))
	for i, l := range f.Lines {
		texts[i] = l.Text
	}
	blob := strings.Join(texts, "\n")
	return ssrfServerFrames.MatchString(blob) || ssrfReqInput.MatchString(blob)
}

func ssrfInput(f diff.File, line int, title string, sev finding.Severity) finding.Input {
	return finding.Input{
		Title:    title,
		Class:    finding.SSRF,
		Path:     f.Path,
		Line:     line,
		CWE:      "CWE-918",
		OWASP:    finding.LabelsFor(finding.SSRF),
		Severity: sev,
	}
}

// SSRF flags server-side client calls with a non-literal URL, and fetched
// bodies echoed back to the caller.
func SSRF(f diff.File) []finding.Finding {
	if l := langOf(f.Path); l == langPython || l == langGo {
		var out []finding.Finding
		for _, line := range ssrfLangLines(f, l) {
			out = add(out, ssrfInput(f, line, "SSRF / open URL fetch", finding.High))
		}
		return out
	}
	if !looksServerSide(f) {
		return nil
	}
	var out []finding.Finding
	allowChecked := false
	for _, line := range f.Lines {
		if IsCommentLine(line.Text) {
			continue
		}
		allowChecked = allowChecked || ssrfAllowCheck.MatchString(line.Text)
		if line.Kind != diff.Added || allowChecked {
			continue
		}
		arg, ok := ssrfURLArg(line.Text)
		if !ok || ssrfFixedTarget(arg) {
			continue
		}
		out = add(out, ssrfInput(f, line.Line, "SSRF / open URL fetch", finding.High))
	}

	fetchedText := false
	for _, line := range f.Added {
		if ssrfResponseText.MatchString(line.Text) {
			fetchedText = true
			break
		}
	}
	if !fetchedText {
		return out
	}
	for _, line := range f.Added {
		if IsCommentLine(line.Text) || !ssrfReturnsHTML.MatchString(line.Text) {
			continue
		}
		out = add(out, ssrfInput(f, line.Line, "Fetched body returned to client", finding.Medium))
	}
	return out
}
