package rules

import (
	"regexp"
	"strings"
)

// strExpr is one source expression split into the text of its string
// literals, the expressions interpolated inside them, and the code outside
// them (every literal replaced by "").
type strExpr struct {
	literal string
	interps []string
	code    string
}

var (
	templateInterp = regexp.MustCompile(`\$\{([^}]*)\}`)
	fstringInterp  = regexp.MustCompile(`\{([^{}]*)\}`)
	phpInterp      = regexp.MustCompile(`\{(\$[^}]*)\}|(\$[A-Za-z_]\w*(?:->\w+|\[[^\]]*\])*)`)
)

// scanExpr splits s into literal text, interpolations and code. php enables
// "$var" interpolation inside double-quoted strings.
func scanExpr(s string, php bool) strExpr {
	var e strExpr
	var code []byte
	var lit strings.Builder
	for i := 0; i < len(s); {
		c := s[i]
		if c != '"' && c != '\'' && c != '`' {
			code = append(code, c)
			i++
			continue
		}
		var fstr bool
		code, fstr = cutStringPrefix(code)
		end, from, to := literalBounds(s, i)
		body := s[from:to]
		switch {
		case c == '`':
			body = e.takeInterps(body, templateInterp)
		case fstr:
			body = e.takeInterps(strings.NewReplacer("{{", "", "}}", "").Replace(body), fstringInterp)
		case php && c == '"':
			body = e.takeInterps(body, phpInterp)
		}
		lit.WriteString(body)
		lit.WriteByte(' ')
		code = append(code, '"', '"')
		i = end
	}
	e.literal = lit.String()
	e.code = string(code)
	return e
}

// takeInterps records every interpolation of re in body and returns body without them.
func (e *strExpr) takeInterps(body string, re *regexp.Regexp) string {
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		for _, g := range m[1:] {
			if g != "" {
				e.interps = append(e.interps, strings.TrimSpace(g))
				break
			}
		}
	}
	return re.ReplaceAllString(body, " ")
}

// cutStringPrefix drops a Python string prefix (f, r, b, u) from the end of
// code and reports whether it makes an f-string.
func cutStringPrefix(code []byte) ([]byte, bool) {
	n := len(code)
	i := n
	for i > 0 && i > n-2 && strings.IndexByte("fFrRbBuU", code[i-1]) >= 0 {
		i--
	}
	if i == n || (i > 0 && (isWordByte(code[i-1]) || code[i-1] == '$' || code[i-1] == '.')) {
		return code, false
	}
	return code[:i], strings.ContainsAny(string(code[i:]), "fF")
}

// literalBounds returns, for the literal opened at s[open], the index just
// past its closing quote and the [from, to) range of its body. Unclosed
// literals run to the end of s.
func literalBounds(s string, open int) (end, from, to int) {
	q := s[open]
	if triple := strings.Repeat(string(q), 3); q != '`' && strings.HasPrefix(s[open:], triple) {
		if k := strings.Index(s[open+3:], triple); k >= 0 {
			return open + 3 + k + 3, open + 3, open + 3 + k
		}
		return len(s), open + 3, len(s)
	}
	for j := open + 1; j < len(s); j++ {
		switch {
		case s[j] == '\\':
			j++
		case s[j] == q:
			return j + 1, open + 1, j
		case q == '`' && s[j] == '$' && j+1 < len(s) && s[j+1] == '{':
			j = closingBrace(s, j+1)
		}
	}
	return len(s), open + 1, len(s)
}

// closingBrace returns the index of the brace closing the one at s[open], or len(s)-1.
func closingBrace(s string, open int) int {
	depth := 0
	for j := open; j < len(s); j++ {
		switch s[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return len(s) - 1
}

// withoutQuoted blanks single- and double-quoted literals, so backticks
// inside SQL strings are not read as PHP shell execution.
func withoutQuoted(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '"' || s[i] == '\'' {
			end, _, _ := literalBounds(s, i)
			sb.WriteString(`""`)
			i = end
			continue
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}

// unquote returns the body of a string literal, or "" for anything else.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if IsStringLiteral(s) {
		return s[1 : len(s)-1]
	}
	return ""
}
