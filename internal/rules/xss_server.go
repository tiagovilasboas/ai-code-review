package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// phpCast is a numeric cast: the value cannot carry markup.
var phpCast = regexp.MustCompile(`^\((?:int|integer|float|double|bool)\)`)

// outputSource is a call whose result is command output, file content or a
// stored record: data fetched with request input is not request input.
var outputSource = regexp.MustCompile(`\b(?:shell_exec|exec|system|passthru|popen|file_get_contents|curl_exec|subprocess\.\w+|os\.popen|requests\.\w+|http\.Get|Output|CombinedOutput|query|execute|fetch\w*)\s*\(`)

// htmlWriter describes how one language writes a string to the HTTP response.
type htmlWriter struct {
	sink       *regexp.Regexp // group 1: the written value, or the call arguments
	escape     *regexp.Regexp // call that makes a value safe for HTML (must end at its opening parenthesis)
	format     *regexp.Regexp // string-formatting call used in assignments, group 1: its arguments
	skip       *regexp.Regexp // sink lines that are not HTML (JSON helpers)
	notHTML    *regexp.Regexp // hunk declares a non-HTML content type: stay silent
	php        bool           // "$var" interpolation, "." concatenation
	bare       bool           // a bare request value is a sink even without markup (echo)
	formatArgs bool           // printf style: the first argument is the format, the rest are values
}

// serverHTMLHits returns the added lines where a handler writes an HTML
// string built from request input to the response.
func serverHTMLHits(f diff.File) []int {
	switch langOf(f.Path) {
	case langJS:
		return expressHTMLHits(f)
	case langPython:
		return pythonWriter.hits(f)
	case langGo:
		return goWriter.hits(f)
	case langJava:
		return javaWriter.hits(f)
	case langPHP:
		return phpWriter.hits(f)
	}
	return nil
}

// templateBypassHit reports an added line that switches auto-escaping off for
// a non-constant value in a template or a type conversion.
func templateBypassHit(p, text string) bool {
	return djangoSafeHit(p, text) || phpTemplateHit(p, text) || javaTemplateHit(p, text) || goTemplateHit(p, text)
}

func (w htmlWriter) hits(f diff.File) []int {
	for _, l := range f.Lines {
		if w.notHTML != nil && w.notHTML.MatchString(l.Text) {
			return nil
		}
	}
	tainted := taintedNames(f.Lines, w.php)
	htmlVars := map[string]bool{}
	var out []int
	for _, l := range f.Lines {
		if IsCommentLine(l.Text) {
			continue
		}
		if a, ok := parseAssignment(l.Text); ok && a.op != "+=" && a.op != ".=" {
			if coveredByCall(a.rhs, w.escape) || outputSource.MatchString(a.rhs) || w.php && phpCast.MatchString(strings.TrimSpace(a.rhs)) {
				for _, n := range a.names {
					delete(tainted, n)
				}
			} else if w.assignedHTML(a.rhs, tainted) {
				for _, n := range a.names {
					htmlVars[n] = true
				}
			}
		}
		if l.Kind != diff.Added || (w.skip != nil && w.skip.MatchString(l.Text)) {
			continue
		}
		if m := w.sink.FindStringSubmatch(l.Text); m != nil && w.writes(splitCallArgs(m[1]), tainted, htmlVars) {
			out = append(out, l.Line)
		}
	}
	return out
}

func (w htmlWriter) assignedHTML(rhs string, tainted map[string]bool) bool {
	if w.format != nil {
		if m := w.format.FindStringSubmatch(rhs); m != nil {
			return w.writes(splitCallArgs(m[1]), tainted, nil)
		}
	}
	return w.exprHasInput(rhs, tainted)
}

// writes reports a sink call whose arguments put request input into HTML,
// directly or through a variable already holding such HTML.
func (w htmlWriter) writes(args []string, tainted, htmlVars map[string]bool) bool {
	if len(args) == 0 {
		return false
	}
	if htmlVars[strings.TrimSpace(args[0])] || w.exprHasInput(args[0], tainted) {
		return true
	}
	if !w.formatArgs || !(w.bare || htmlMarkup.MatchString(scanExpr(args[0], w.php).literal)) {
		return false
	}
	for _, a := range args[1:] {
		if w.unescapedInput(a, tainted) {
			return true
		}
	}
	return false
}

// exprHasInput reports an expression that is HTML text (its literal parts
// hold markup, or any value counts when bare) with request input mixed in.
func (w htmlWriter) exprHasInput(expr string, tainted map[string]bool) bool {
	e := scanExpr(expr, w.php)
	if !w.bare && !htmlMarkup.MatchString(e.literal) {
		return false
	}
	sep := "+"
	if w.php {
		sep = "."
	}
	pieces := append([]string{}, e.interps...)
	pieces = append(pieces, strings.Split(e.code, sep)...)
	for _, p := range pieces {
		if w.unescapedInput(p, tainted) {
			return true
		}
	}
	return false
}

func (w htmlWriter) unescapedInput(p string, tainted map[string]bool) bool {
	p = strings.TrimSpace(p)
	return p != "" && !coveredByCall(p, w.escape) && !phpCast.MatchString(p) && reachesRequest(p, tainted, w.php)
}
