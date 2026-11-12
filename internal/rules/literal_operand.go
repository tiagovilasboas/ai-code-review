package rules

import (
	"regexp"
	"strings"
)

// pureLiteralRe matches a whole expression that is exactly one quoted string
// with no interpolation, optionally followed by closing punctuation from the
// surrounding call or JSX.
var pureLiteralRe = regexp.MustCompile("^\\s*(?:'[^'\\\\\\n]*'|\"[^\"\\\\\\n]*\"|`[^`$\\\\\\n]*`)[;,)}\\s/>`]*$")

// isPureLiteral reports whether expr is a single constant string literal.
func isPureLiteral(expr string) bool {
	return pureLiteralRe.MatchString(expr)
}

// coveredByCall reports whether expr starts with a call matching callRe (the
// pattern must end at the opening parenthesis) and that call's closing
// parenthesis ends the expression, so nothing is concatenated around it.
func coveredByCall(expr string, callRe *regexp.Regexp) bool {
	expr = strings.TrimSpace(expr)
	loc := callRe.FindStringIndex(expr)
	if loc == nil || loc[0] != 0 {
		return false
	}
	depth := 1
	for i := loc[1]; i < len(expr); i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				rest := strings.TrimSpace(expr[i+1:])
				return rest == "" || strings.ContainsRune(";,})/>`", rune(rest[0])) || strings.HasPrefix(rest, "?>")
			}
		}
	}
	return false
}
