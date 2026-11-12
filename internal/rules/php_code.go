package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var phpHeredoc = regexp.MustCompile(`<<<\s*(?:'(\w+)'|"(\w+)"|(\w+))\s*$`)

// phpCodeOnly returns one string per line with the body of every string
// literal and comment blanked to spaces (same length, quotes kept). A quoted
// string, heredoc or block comment that spans several added lines stays open
// across them, so SQL text on a continuation line is never read as PHP code.
// Text inside backticks is code (shell execution) and is left as written.
func phpCodeOnly(lines []diff.Line) []string {
	out := make([]string, len(lines))
	var state byte // 0 code, '"' or '\'' open literal, 'h' heredoc, '*' block comment
	label := ""
	for i, l := range lines {
		if i > 0 && l.Line != lines[i-1].Line+1 {
			state, label = 0, ""
		}
		out[i], state, label = phpBlankLine(l.Text, state, label)
	}
	return out
}

func phpBlankLine(s string, state byte, label string) (string, byte, string) {
	b := []byte(s)
	blank := func(from, to int) {
		for k := from; k < to && k < len(b); k++ {
			b[k] = ' '
		}
	}
	i := 0
	if state == 'h' {
		if t := strings.TrimLeft(s, " \t"); strings.HasPrefix(t, label) && (len(t) == len(label) || !isWordByte(t[len(label)])) {
			return s, 0, ""
		}
		blank(0, len(b))
		return string(b), 'h', label
	}
	for i < len(b) {
		switch {
		case state == '*':
			end := strings.Index(s[i:], "*/")
			if end < 0 {
				blank(i, len(b))
				return string(b), state, label
			}
			blank(i, i+end+2)
			i, state = i+end+2, 0
		case state == '"' || state == '\'':
			switch {
			case b[i] == '\\':
				blank(i, i+2)
				i += 2
			case b[i] == state:
				state = 0
				i++
			default:
				b[i] = ' '
				i++
			}
		case b[i] == '"' || b[i] == '\'':
			state = b[i]
			i++
		case b[i] == '`':
			end := strings.IndexByte(s[i+1:], '`')
			if end < 0 {
				return string(b), 0, ""
			}
			i += end + 2
		case strings.HasPrefix(s[i:], "/*"):
			state = '*'
			blank(i, i+2)
			i += 2
		case strings.HasPrefix(s[i:], "//") || (b[i] == '#' && !strings.HasPrefix(s[i:], "#[")):
			blank(i, len(b))
			return string(b), 0, ""
		case strings.HasPrefix(s[i:], "<<<"):
			if m := phpHeredoc.FindStringSubmatch(s[i:]); m != nil {
				return string(b), 'h', m[1] + m[2] + m[3]
			}
			i += 3
		default:
			i++
		}
	}
	return string(b), state, label
}
