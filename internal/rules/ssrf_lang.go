package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// Python and Go outbound HTTP calls. Unlike the JS rule, these require the
// destination to trace back to request input, because the same call shapes
// (requests.get, http.Get) are everywhere in server code with fixed URLs.
var (
	ssrfPyCall = regexp.MustCompile(`\b(?:requests|httpx|urllib\.request|\w*(?i:client|http)\w*)\.(get|post|put|patch|delete|head|options|request|stream|urlopen|urlretrieve)\s*\(\s*(.*)$`)
	ssrfPyBare = regexp.MustCompile(`(?:^|[^.\w])(urlopen)\s*\(\s*(.*)$`)
	ssrfGoCall = regexp.MustCompile(`\bhttp\.(Get|Head|Post|PostForm|NewRequest|NewRequestWithContext)\s*\(\s*(.*)$`)
	ssrfGoClnt = regexp.MustCompile(`\b\w*(?i:client)\w*\.(Get|Head|Post|PostForm)\s*\(\s*(.*)$`)

	// ssrfGoResty: resty-style chains, client.R().SetX(..).Get(url).
	ssrfGoResty = regexp.MustCompile(`\.R\(\)(?:\.\w+\([^()]*\))*\.(Get|Post|Put|Patch|Delete|Head)\s*\(\s*(.*)$`)

	ssrfPyPrefix = regexp.MustCompile(`^[fFrRbBuU]{1,2}['"]`)
	// ssrfLangBase: the URL starts from deployment configuration.
	ssrfLangBase = regexp.MustCompile(`^(?:os\.environ|os\.getenv|os\.Getenv|settings\.|current_app\.config|app\.config|config\.|cfg\.|System\.get(?:env|Property)\(|env\.getProperty\(|(?:config|env|getenv)\()`)
	// ssrfLangAllow recognizes a membership check against an allowlist.
	ssrfLangAllow = regexp.MustCompile(`(?i)\ballow\w*\s*\[|\b(?:slices\.)?Contains\(\s*\w*allow|\bin\s+\w*allow\w*|(?i:\w*allow\w*)\.contains\(|\bin_array\([^;]*allow`)
	// ssrfBodyDecode binds the request body to a variable (Go).
	ssrfBodyDecode = regexp.MustCompile(`(?:\b(?:r|req|request)\.Body\b|\bc\.Request\.Body\b).*?(?:Decode|Unmarshal)\(.*?&(\w+)|\bc\.\w*Bind\w*\(\s*&(\w+)`)
)

// ssrfLangDest returns the destination expression of a Python or Go
// outbound call on text.
func ssrfLangDest(l lang, text string) (string, bool) {
	var m []string
	switch l {
	case langPython:
		if m = ssrfPyCall.FindStringSubmatch(text); m == nil {
			m = ssrfPyBare.FindStringSubmatch(text)
		}
	case langJava:
		m = ssrfJavaCall.FindStringSubmatch(text)
	case langPHP:
		return ssrfPHPDest(text)
	default:
		if m = ssrfGoCall.FindStringSubmatch(text); m == nil {
			if m = ssrfGoClnt.FindStringSubmatch(text); m == nil {
				m = ssrfGoResty.FindStringSubmatch(text)
			}
		}
	}
	if m == nil {
		return "", false
	}
	args := splitCallArgs(m[2])
	idx := 0
	switch m[1] {
	case "request", "stream", "NewRequest":
		idx = 1
	case "NewRequestWithContext":
		idx = 2
	}
	if idx >= len(args) {
		return "", false
	}
	if l == langJava && strings.Contains(args[idx], "->") {
		return "", false // WebClient uriBuilder lambda: path and query only
	}
	return ssrfStripPrefix(args[idx]), true
}

// splitCallArgs splits the text after an opening parenthesis into the call's
// top-level arguments, up to the matching closing parenthesis.
func splitCallArgs(s string) []string {
	var args []string
	depth, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'', '"', '`':
			i = skipQuoted(s, i)
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth == 0 {
				return appendCallArg(args, s[start:i])
			}
			depth--
		case ',':
			if depth == 0 {
				args = appendCallArg(args, s[start:i])
				start = i + 1
			}
		}
	}
	return appendCallArg(args, s[start:])
}

func appendCallArg(args []string, a string) []string {
	if a = strings.TrimSpace(a); a == "" {
		return args
	}
	return append(args, a)
}

// ssrfStripPrefix drops a Python string prefix (f, r, b) so the literal
// checks see the quote first.
func ssrfStripPrefix(s string) string {
	return ssrfPyPrefix.ReplaceAllStringFunc(strings.TrimSpace(s), func(m string) string { return m[len(m)-1:] })
}

func ssrfLangFixed(arg string) bool {
	return ssrfFixedTarget(arg) || ssrfLangBase.MatchString(arg)
}

// ssrfRequestTaint lists the variables holding request input. Assignments of
// a fixed-host or configured URL are left out, so a request value used only
// as a path segment does not taint the whole URL.
func ssrfRequestTaint(lines []diff.Line, php bool) map[string]bool {
	var kept []diff.Line
	var bound []string
	for _, l := range lines {
		if m := ssrfBodyDecode.FindStringSubmatch(l.Text); m != nil {
			bound = append(bound, m[1]+m[2])
		}
		if a, ok := parseAssignment(l.Text); ok && ssrfLangFixed(ssrfStripPrefix(a.rhs)) {
			continue
		}
		kept = append(kept, l)
	}
	tainted := taintedNames(kept, php)
	for _, n := range bound {
		tainted[n] = true
	}
	return tainted
}

// ssrfLangLines returns the added lines of a Python, Go, Java or PHP file that call out
// to a destination taken from request input with no allowlist check before.
func ssrfLangLines(f diff.File, l lang) []int {
	php := l == langPHP
	tainted := ssrfRequestTaint(f.Lines, php)
	var out []int
	allowChecked := false
	for _, line := range f.Lines {
		if IsCommentLine(line.Text) {
			continue
		}
		allowChecked = allowChecked || ssrfAllowCheck.MatchString(line.Text) || ssrfLangAllow.MatchString(line.Text)
		if line.Kind != diff.Added || allowChecked {
			continue
		}
		dest, ok := ssrfLangDest(l, line.Text)
		if !ok || ssrfLangFixed(dest) || !reachesRequest(dest, tainted, php) {
			continue
		}
		out = append(out, line.Line)
	}
	return out
}
