package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// callArgs returns the top-level arguments of the call whose '(' is at
// text[open]. Arguments of a call that does not close in text are kept.
func callArgs(text string, open int) []string {
	var args []string
	depth, start := 0, open+1
	for i := open; i < len(text); {
		switch text[i] {
		case '"', '\'', '`':
			i, _, _ = literalBounds(text, i)
			continue
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth == 0 {
				return appendArg(args, text[start:i])
			}
		case ',':
			if depth == 1 {
				args = appendArg(args, text[start:i])
				start = i + 1
			}
		}
		i++
	}
	return appendArg(args, text[start:])
}

func appendArg(args []string, a string) []string {
	if a = strings.TrimSpace(a); a != "" {
		return append(args, a)
	}
	return args
}

// maxCallLines bounds how far a call may span below its sink line.
const maxCallLines = 12

// callFrom returns lines[index] from column col, joined with the following
// lines of the same hunk, so that multi-line calls can be read whole.
func callFrom(lines []diff.Line, index, col int) string {
	var sb strings.Builder
	sb.WriteString(lines[index].Text[col:])
	for i := index + 1; i < len(lines) && i <= index+maxCallLines; i++ {
		if lines[i].Line != lines[i-1].Line+1 {
			break
		}
		sb.WriteByte('\n')
		sb.WriteString(lines[i].Text)
	}
	return sb.String()
}

// hunkAround returns the contiguous hunk lines that contain lines[index].
func hunkAround(lines []diff.Line, index int) []diff.Line {
	start, end := index, index
	for start > 0 && lines[start-1].Line == lines[start].Line-1 {
		start--
	}
	for end+1 < len(lines) && lines[end+1].Line == lines[end].Line+1 {
		end++
	}
	return lines[start : end+1]
}

// anyLineMatches reports whether any visible line matches re.
func anyLineMatches(lines []diff.Line, re *regexp.Regexp) bool {
	for _, l := range lines {
		if re.MatchString(l.Text) {
			return true
		}
	}
	return false
}
