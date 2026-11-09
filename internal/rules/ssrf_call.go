package rules

import (
	"regexp"
	"strings"
)

var (
	// ssrfClientCall captures what follows the opening parenthesis of an
	// outbound HTTP client call.
	ssrfClientCall = regexp.MustCompile(`\b(?:fetch|axios(?:\.(?:get|post|put|patch|delete|head|request))?|got(?:\.(?:get|post|put|patch|delete|head))?|https?\.(?:get|request))\s*\(\s*(.*)$`)
	// ssrfConfigURL finds the url key of a config-object call such as axios({ url }).
	ssrfConfigURL = regexp.MustCompile(`\burl\s*:\s*`)
)

// ssrfURLArg returns the expression naming the destination of an outbound
// client call on text: the first argument, or the url key of a config object.
func ssrfURLArg(text string) (string, bool) {
	m := ssrfClientCall.FindStringSubmatch(text)
	if m == nil {
		return "", false
	}
	rest := m[1]
	if strings.HasPrefix(rest, "{") {
		loc := ssrfConfigURL.FindStringIndex(rest)
		if loc == nil {
			return "", false
		}
		rest = rest[loc[1]:]
	}
	arg := firstExpr(rest)
	return arg, arg != ""
}

// firstExpr returns the leading expression of s, ending at a top-level comma
// or an unmatched closing bracket; quoted strings are skipped whole.
func firstExpr(s string) string {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\'', '"', '`':
			i = skipQuoted(s, i)
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth == 0 {
				return strings.TrimSpace(s[:i])
			}
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(s[:i])
			}
		}
	}
	return strings.TrimSpace(s)
}

// skipQuoted returns the index of the quote closing the one opened at i.
func skipQuoted(s string, i int) int {
	q := s[i]
	for j := i + 1; j < len(s); j++ {
		if s[j] == '\\' {
			j++
		} else if s[j] == q {
			return j
		}
	}
	return len(s)
}
