package rules

import "regexp"

var (
	goWriter = htmlWriter{
		// Writes to the response writer (w, rw, res, c.Writer, gin c.Data as HTML).
		sink: regexp.MustCompile(`\b(?:fmt\.Fprint(?:f|ln)?\(\s*(?:w|rw|wr|res|resp|writer|c\.Writer|ctx\.Writer)\s*,` +
			`|io\.WriteString\(\s*(?:w|rw|wr|res|resp|writer|c\.Writer|ctx\.Writer)\s*,` +
			`|(?:w|rw|wr|res|resp|writer|c\.Writer|ctx\.Writer)\.Write\(\s*\[\]byte\(` +
			`|c\.Data\(\s*[\w.]+\s*,\s*"text/html[^"]*"\s*,\s*\[\]byte\(` +
			// echo: c.HTML(code, html) and c.HTMLBlob(code, []byte(html)).
			`|c\.HTML\(\s*[\w.]+\s*,` +
			`|c\.HTMLBlob\(\s*[\w.]+\s*,\s*\[\]byte\()(.*)$`),
		escape:     regexp.MustCompile(`(?:html\.EscapeString|template\.\w*Escape\w*|\w+\.Sanitize\w*|url\.QueryEscape|strconv\.\w+)\s*\(`),
		format:     regexp.MustCompile(`\bfmt\.Sprintf\(\s*(.*)$`),
		notHTML:    regexp.MustCompile(`(?i)Content-Type["']?\s*,\s*"(?:application/json|text/plain)`),
		formatArgs: true,
	}
	// html/template conversions that switch contextual escaping off.
	goTemplateBypass = regexp.MustCompile(`\btemplate\.(?:HTML|JS|CSS|HTMLAttr)\(\s*(.*)$`)
)

// goTemplateHit reports template.HTML(x) (and its siblings) over a value that
// is neither a constant nor sanitized in place.
func goTemplateHit(p, text string) bool {
	if langOf(p) != langGo {
		return false
	}
	m := goTemplateBypass.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	args := splitCallArgs(m[1])
	return len(args) > 0 && !isPureLiteral(args[0]) && !coveredByCall(args[0], goWriter.escape)
}
