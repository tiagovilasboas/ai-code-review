package rules

import "strings"

func isWordByte(b byte) bool {
	return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// braceNet returns opening minus closing braces, ignoring string literals.
func braceNet(text string) int {
	n := 0
	for _, ch := range []byte(stripStrings(text)) {
		switch ch {
		case '{':
			n++
		case '}':
			n--
		}
	}
	return n
}

// stripStrings replaces string literals with "" so braces or parens inside
// them do not skew counting. It ports the JS regex that matches a quote,
// escape pairs or non-quote characters, then the same quote (backreference
// and lookahead are unavailable in RE2), including its backtracking.
func stripStrings(text string) string {
	var sb strings.Builder
	for i := 0; i < len(text); {
		c := text[i]
		if c == '"' || c == '\'' || c == '`' {
			if end, ok := stringEnd(text, i, c); ok {
				sb.WriteString(`""`)
				i = end + 1
				continue
			}
		}
		sb.WriteByte(c)
		i++
	}
	return sb.String()
}

// stringEnd finds the quote closing the literal opened at open, trying an
// escape pair before a plain character, like the JS alternation.
func stringEnd(text string, open int, q byte) (int, bool) {
	failed := make(map[int]bool)
	var try func(pos int) (int, bool)
	try = func(pos int) (int, bool) {
		if pos >= len(text) || failed[pos] {
			return 0, false
		}
		c := text[pos]
		if c == '\\' && pos+1 < len(text) && text[pos+1] != '\n' {
			if end, ok := try(pos + 2); ok {
				return end, true
			}
		}
		if c == q {
			return pos, true
		}
		if c != '\n' {
			if end, ok := try(pos + 1); ok {
				return end, true
			}
		}
		failed[pos] = true
		return 0, false
	}
	return try(open + 1)
}
