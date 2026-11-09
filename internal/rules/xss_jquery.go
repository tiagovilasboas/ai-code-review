package rules

import "regexp"

// xssJQueryHTML matches jQuery's .html(value) setter; the .html() getter has
// no argument and does not match.
var xssJQueryHTML = regexp.MustCompile(`(?:\$\(|\bjQuery\(|\$\w+)[^\n]*?\.html\(\s*([^)\s][^\n]*)$`)

// jqueryHTMLWrite reports a .html() write whose value is neither constant nor sanitized.
func jqueryHTMLWrite(text string) bool {
	m := xssJQueryHTML.FindStringSubmatch(text)
	return m != nil && !constOrSanitized(m[1])
}
