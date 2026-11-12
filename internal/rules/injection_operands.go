package rules

import (
	"regexp"
	"strings"
)

var (
	identChain = regexp.MustCompile(`\$?[A-Za-z_]\w*(?:\s*(?:\?\.|\.|->|::)\s*\$?[A-Za-z_]\w*)*`)
	buildOp    = regexp.MustCompile(`\+|%|\bSprintf\s*\(|\bformat\s*\(|\.concat\s*\(`)
	phpConcat  = regexp.MustCompile(`""\s*\.|\.\s*""|\s\.\s|\.\$`)
	constName  = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	numberLit  = regexp.MustCompile(`^-?\d+(?:\.\d+)?$`)
)

// notOperands are keywords, types and string builders: none carries data.
var notOperands = map[string]bool{
	"true": true, "false": true, "null": true, "nil": true, "None": true, "True": true,
	"False": true, "undefined": true, "new": true, "this": true, "self": true,
	"String": true, "Integer": true, "Long": true, "Object": true, "List": true,
	"str": true, "int": true, "typeof": true, "await": true, "return": true,
	"Sprintf": true, "format": true, "text": true, "sql": true, "concat": true,
	"of": true, "asList": true,
}

// isLiteralOperand reports whether an interpolated or concatenated value is
// constant: a string or number literal, an UPPER_CASE constant or a keyword.
func isLiteralOperand(s string) bool {
	s = strings.TrimSpace(s)
	if k := strings.LastIndex(s, "::"); k >= 0 && !strings.Contains(s, "(") {
		s = s[k+2:] // self::TABLE, static::TABLE, Foo::BAR are class constants
	}
	return IsStringLiteral(s) || numberLit.MatchString(s) || constName.MatchString(s) || notOperands[s]
}

// operands lists the non-literal values in code, call names included.
func operands(code string) []string {
	var out []string
	for _, loc := range identChain.FindAllStringIndex(code, -1) {
		if loc[0] > 0 && isWordByte(code[loc[0]-1]) {
			continue // tail of a number such as 1e5
		}
		name := code[loc[0]:loc[1]]
		last := strings.TrimSpace(name[strings.LastIndexAny(name, ".>:")+1:])
		if isLiteralOperand(name) || notOperands[last] {
			continue
		}
		out = append(out, name)
	}
	return out
}

// nonLiteralOperands lists every value of e that is not a constant.
func (e strExpr) nonLiteralOperands() []string {
	return append(e.interpOperands(), operands(e.code)...)
}

func (e strExpr) interpOperands() []string {
	var out []string
	for _, in := range e.interps {
		if !isLiteralOperand(in) {
			out = append(out, in)
		}
	}
	return out
}

// dynamicOperands lists the non-constant values e builds into a string:
// interpolations, plus code values when e concatenates or formats.
func (e strExpr) dynamicOperands(php bool) []string {
	if buildOp.MatchString(e.code) || (php && phpConcat.MatchString(e.code)) {
		return e.nonLiteralOperands()
	}
	return e.interpOperands()
}

// mentions reports whether code uses the variable name as a whole word,
// not as a member of another object.
func mentions(code, name string) bool {
	for from := 0; ; {
		k := strings.Index(code[from:], name)
		if k < 0 {
			return false
		}
		at, end := from+k, from+k+len(name)
		before := at == 0 || !(isWordByte(code[at-1]) || strings.IndexByte(".$>", code[at-1]) >= 0)
		if before && (end == len(code) || !isWordByte(code[end])) {
			return true
		}
		from = at + 1
	}
}

func mentionsAny(code string, names map[string]bool) bool {
	for n := range names {
		if mentions(code, n) {
			return true
		}
	}
	return false
}
