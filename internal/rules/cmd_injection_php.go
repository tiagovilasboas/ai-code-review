package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// escapeArgCall opens an escapeshellarg( call. escapeshellcmd is deliberately
// absent: it escapes shell metacharacters of a whole command, not the
// arguments, so a value inside it can still inject options and arguments.
var escapeArgCall = regexp.MustCompile(`\bescapeshellarg\s*\(`)

// phpShellCall: a variable reaches the command unless escapeshellarg wraps it
// or it only holds fully escaped text.
func phpShellCall(args []string, c cmdCtx) bool {
	if len(args) == 0 {
		return false
	}
	for _, op := range scanExpr(stripEscapedArgs(args[0]), true).nonLiteralOperands() {
		if !c.safeCmd[strings.TrimSpace(op)] {
			return true
		}
	}
	return false
}

// stripEscapedArgs replaces each escapeshellarg(...) call, nested calls
// inside it included, with an empty string literal.
func stripEscapedArgs(s string) string {
	for {
		loc := escapeArgCall.FindStringIndex(s)
		if loc == nil {
			return s
		}
		s = s[:loc[0]] + `""` + s[callEnd(s, loc[1]-1):]
	}
}

// callEnd returns the index just past the parenthesis closing s[open].
func callEnd(s string, open int) int {
	depth := 0
	for i := open; i < len(s); {
		switch s[i] {
		case '"', '\'':
			i, _, _ = literalBounds(s, i)
			continue
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
		i++
	}
	return len(s)
}

// phpEscapedVars maps each variable to whether it holds only literal text and
// escapeshellarg results, so that running it adds no unescaped value.
func phpEscapedVars(lines []diff.Line) map[string]bool {
	safe := map[string]bool{}
	for _, l := range lines {
		if IsCommentLine(l.Text) {
			continue
		}
		a, ok := parseAssignment(l.Text)
		if !ok || len(a.names) != 1 {
			continue
		}
		clean := onlyEscaped(a.rhs, safe)
		if a.appends() {
			clean = clean && safe[a.names[0]]
		}
		safe[a.names[0]] = clean
	}
	return safe
}

func onlyEscaped(expr string, safe map[string]bool) bool {
	for _, op := range scanExpr(stripEscapedArgs(expr), true).nonLiteralOperands() {
		if !safe[strings.TrimSpace(op)] {
			return false
		}
	}
	return true
}
