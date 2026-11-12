package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	// expressSend matches res.send/write/end (also after .status(n) and
	// .set(..) chains) and reply.send; group 1 is the chain, group 2 the call.
	expressSend = regexp.MustCompile(`\b(?:res|reply)((?:\.\w+\([^()]*\))*)\.(?:send|write|end)\s*\((.*)$`)
	// expressNotHTML: the chain declares a non-HTML content type.
	expressNotHTML = regexp.MustCompile(`text/plain|json|javascript|octet-stream`)
	htmlMarkup     = regexp.MustCompile(`<[A-Za-z!/]`)
	htmlEscapeCall = regexp.MustCompile(`(?:\b(?:escapeHtml|escapeHTML|escape|encode|sanitizeHtml|sanitizeHTML|sanitize|xss)|\bhe\.(?:encode|escape)|\bvalidator\.escape|\b_\.escape|\bDOMPurify\.sanitize)\s*\(`)
	expressIdent   = regexp.MustCompile(`^[A-Za-z_$][\w$]*$`)
)

// expressHTMLHits returns the added lines of a JS/TS file that write an HTML
// string built from request input to an Express/Fastify response.
func expressHTMLHits(f diff.File) []int {
	if langOf(f.Path) != langJS {
		return nil
	}
	tainted := taintedNames(f.Lines, false)
	htmlVars := map[string]bool{}
	var out []int
	for _, line := range f.Lines {
		if IsCommentLine(line.Text) {
			continue
		}
		if a, ok := parseAssignment(line.Text); ok && a.op == "=" && htmlWithRequestInput(a.rhs, tainted) {
			for _, n := range a.names {
				htmlVars[n] = true
			}
		}
		if line.Kind != diff.Added {
			continue
		}
		m := expressSend.FindStringSubmatch(line.Text)
		if m == nil || expressNotHTML.MatchString(m[1]) {
			continue
		}
		args := splitCallArgs(m[2])
		if len(args) == 0 {
			continue
		}
		if htmlWithRequestInput(args[0], tainted) || (expressIdent.MatchString(args[0]) && htmlVars[args[0]]) {
			out = append(out, line.Line)
		}
	}
	return out
}

// htmlWithRequestInput reports an expression that is HTML text (its literal
// parts hold markup) with a request-derived value mixed in unescaped.
func htmlWithRequestInput(expr string, tainted map[string]bool) bool {
	return htmlWriter{escape: htmlEscapeCall}.exprHasInput(expr, tainted)
}
