package rules

import "regexp"

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

var (
	xssAssignRHS   = regexp.MustCompile(`\.(?:inner|outer)HTML\s*=\s*(.*)$`)
	xssHTMLValue   = regexp.MustCompile(`__html\s*:\s*(.*)$`)
	xssVHTMLValue  = regexp.MustCompile(`\bv-html\s*=\s*["'](.*)$`)
	xssCallArgs    = regexp.MustCompile(`\b(?:insertAdjacentHTML|document\.write|unsafeHTML|bypassSecurityTrust\w+)\s*\((.*)$`)
	xssAdjacentPos = regexp.MustCompile(`^\s*['"]\w+['"]\s*,\s*`)
	xssIdentPath   = regexp.MustCompile(`^([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)*)\s*\}`)
)

// sinkHit reports whether text has an HTML sink whose value is neither a
// constant nor sanitized in place.
func (c xssContext) sinkHit(text string) bool {
	for _, re := range xssSinks {
		if re.MatchString(text) {
			return !c.valueIsSafe(text)
		}
	}
	return false
}

// valueIsSafe inspects the value written to the sink for constants and sanitizers.
func (c xssContext) valueIsSafe(text string) bool {
	if m := xssAssignRHS.FindStringSubmatch(text); m != nil {
		return constOrSanitized(m[1])
	}
	if m := xssHTMLValue.FindStringSubmatch(text); m != nil {
		if constOrSanitized(m[1]) {
			return true
		}
		id := xssIdentPath.FindStringSubmatch(m[1])
		return id != nil && c.sanitized[id[1]]
	}
	if m := xssVHTMLValue.FindStringSubmatch(text); m != nil {
		return constOrSanitized(m[1])
	}
	if m := xssCallArgs.FindStringSubmatch(text); m != nil {
		return constOrSanitized(xssAdjacentPos.ReplaceAllString(m[1], ""))
	}
	return false
}

func constOrSanitized(expr string) bool {
	return isPureLiteral(expr) || coveredByCall(expr, xssSanitizerCall)
}
