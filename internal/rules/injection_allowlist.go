package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// An allowlisted value is request input that the hunk proves is one of a
// fixed set of safe values before it reaches a sink: it is replaced by a
// constant, or the handler rejects it unless it is a known member. Injection
// rules treat such a value like a constant.

const (
	allowIdent = `[\w$][\w$.]*`
	allowLit   = `(?:'[^']*'|"[^"]*"|[A-Z][A-Z0-9_]*|-?\d+(?:\.\d+)?|null|nil|None)`
)

var (
	// notMember matches a negated membership test and captures the value tested.
	notMember = []*regexp.Regexp{
		regexp.MustCompile(`!\s*[\w$.]+\.(?:includes|has|contains|containsKey)\(\s*(` + allowIdent + `)\s*\)`),
		regexp.MustCompile(`!\s*in_array\(\s*(` + allowIdent + `)\s*,`),
		regexp.MustCompile(`(` + allowIdent + `)\s+not\s+in\s+[\w.]+`),
		regexp.MustCompile(`!\s*[\w$.]+\[\s*(` + allowIdent + `)\s*\]`),
		regexp.MustCompile(`\[\s*(` + allowIdent + `)\s*\]\s*;\s*!\w+`),
		regexp.MustCompile(`\.indexOf\(\s*(` + allowIdent + `)\s*\)\s*(?:<\s*0|===?\s*-1)`),
	}
	// notNameRegex matches "!/re/.test(v)" and captures the pattern and the value.
	notNameRegex = regexp.MustCompile(`!\s*/((?:\\.|\[(?:\\.|[^\]\\])*\]|[^/\\\n\[])+)/[a-z]*\.test\(\s*(` + allowIdent + `)\s*\)`)
	rejectWord   = regexp.MustCompile(`\b(?:return|throw|raise|abort|die|exit|continue)\b|\.status\(|\.sendStatus\(|\bhttp\.Error\(`)

	// pickAllowed matches an expression that yields a request value only when
	// it is a member of a collection (groups 1 and 2 must name the same value).
	pickAllowed = []*regexp.Regexp{
		regexp.MustCompile(`^[\w$.]+\.(?:includes|has|contains)\(\s*(` + allowIdent + `)\s*\)\s*\?\s*(` + allowIdent + `)\s*:\s*` + allowLit + `$`),
		regexp.MustCompile(`^in_array\(\s*(` + allowIdent + `)\s*,[^)]*\)\s*\?\s*(` + allowIdent + `)\s*:\s*` + allowLit + `$`),
		regexp.MustCompile(`^(` + allowIdent + `)\s+if\s+(` + allowIdent + `)\s+in\s+[\w.]+\s+else\s+` + allowLit + `$`),
	}
	// mapAllowed is a lookup of the request value in a map, with a constant default.
	mapAllowed = regexp.MustCompile(`^(\$?[A-Za-z_][\w.]*)\[[^\]]+\]\s*(?:\?\?|\|\|)\s*` + allowLit + `$`)
	mapGetOr   = regexp.MustCompile(`^(\$?[A-Za-z_][\w.]*)\.get\([^,()]+,\s*` + allowLit + `\)$`)
)

// allowlistedValues returns the names and request expressions that the lines
// prove allowlisted: assigned from an allowlist pick, or tested with a
// rejecting membership check.
func allowlistedValues(lines []diff.Line) map[string]bool {
	safe := map[string]bool{}
	for i, l := range lines {
		if IsCommentLine(l.Text) {
			continue
		}
		if a, ok := parseAssignment(l.Text); ok && a.op == "=" && len(a.names) == 1 && allowedExpr(a.rhs) {
			safe[a.names[0]] = true
		}
		for _, v := range rejectedUnlessAllowed(lines, i) {
			safe[v] = true
		}
	}
	return safe
}

// allowedExpr reports whether expr picks a member of a collection or a
// constant default for a request value.
func allowedExpr(expr string) bool {
	expr = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(expr), ";"))
	for _, re := range pickAllowed {
		if m := re.FindStringSubmatch(expr); m != nil && m[1] == m[2] {
			return true
		}
	}
	for _, re := range []*regexp.Regexp{mapAllowed, mapGetOr} {
		if m := re.FindStringSubmatch(expr); m != nil && !requestInput.MatchString(m[1]) {
			return true
		}
	}
	return false
}

// rejectedUnlessAllowed returns the values that lines[i] tests for set
// membership and whose failure rejects the request.
func rejectedUnlessAllowed(lines []diff.Line, i int) []string {
	text := lines[i].Text
	var out []string
	for _, re := range notMember {
		for _, m := range re.FindAllStringSubmatchIndex(text, -1) {
			if rejects(lines, i, text[m[1]:]) {
				out = append(out, text[m[2]:m[3]])
			}
		}
	}
	for _, m := range notNameRegex.FindAllStringSubmatchIndex(text, -1) {
		if nameOnlyRegex(text[m[2]:m[3]]) && rejects(lines, i, text[m[1]:]) {
			out = append(out, text[m[4]:m[5]])
		}
	}
	return out
}

// rejects reports whether the body of the check on lines[i] stops the request.
// tail is the rest of the check line after the tested condition.
func rejects(lines []diff.Line, i int, tail string) bool {
	tail = strings.TrimSpace(tail)
	if rejectWord.MatchString(tail) {
		return true
	}
	span := 0
	switch {
	case strings.HasSuffix(tail, "{"), strings.HasSuffix(tail, ":"):
		span = 2
	case strings.HasSuffix(tail, ")"):
		span = 1
	}
	for k := 1; k <= span && i+k < len(lines); k++ {
		if lines[i+k].Line != lines[i].Line+k {
			break
		}
		body := strings.TrimSpace(lines[i+k].Text)
		if strings.HasPrefix(body, "}") || strings.HasPrefix(body, "else") {
			break
		}
		if rejectWord.MatchString(body) {
			return true
		}
	}
	return false
}

// nameOnlyRegex reports whether a JS regex body, tested on a whole value,
// admits only plain names: anchored, with no dot, slash or negated class.
func nameOnlyRegex(body string) bool {
	if !strings.HasPrefix(body, "^") || !strings.HasSuffix(body, "$") {
		return false
	}
	return !strings.ContainsAny(body, "./") && !strings.Contains(body, "[^") &&
		!strings.Contains(body, `\S`) && !strings.Contains(body, `\W`) && !strings.Contains(body, `\D`)
}

// dropAllowlisted removes the allowlisted values from a list of operands.
func dropAllowlisted(ops []string, safe map[string]bool) []string {
	var out []string
	for _, op := range ops {
		if t := strings.TrimSpace(op); !safe[t] && !allowedExpr(t) {
			out = append(out, op)
		}
	}
	return out
}

// maskAllowlisted replaces every allowlisted value in code by a constant.
func maskAllowlisted(code string, safe map[string]bool) string {
	for name := range safe {
		var sb strings.Builder
		for from := 0; ; {
			k := strings.Index(code[from:], name)
			if k < 0 {
				sb.WriteString(code[from:])
				break
			}
			at, end := from+k, from+k+len(name)
			startOK := at == 0 || !(isWordByte(code[at-1]) || strings.IndexByte(".$>", code[at-1]) >= 0)
			if startOK && (end == len(code) || !isWordByte(code[end])) {
				sb.WriteString(code[from:at] + `""`)
			} else {
				sb.WriteString(code[from:end])
			}
			from = end
		}
		code = sb.String()
	}
	return code
}
