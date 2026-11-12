package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	goHTMLBuffer = regexp.MustCompile(`\btemplate\.(?:HTML|JS|CSS|HTMLAttr)\(\s*(?:string\(\s*(\w+)\.Bytes\(\)\s*\)|(\w+)\.(?:String|Bytes)\(\)|(\w+))\s*\)`)
	goFuncStart  = regexp.MustCompile(`^\s*func\b`)
	// A string constant declared in the function: `name = "<span>"`, alone
	// in a const block or after const.
	goLocalStrConst = regexp.MustCompile("^\\s*(?:const\\s+)?(\\w+)\\s*=\\s*(?:\"[^\"]*\"|`[^`]*`)\\s*$")
)

// goBufferEscaped reports template.HTML(buf.String()) (or template.HTML(buf))
// where every write to buf in the same function is a literal or went through
// an HTML escaper. A buffer filled any other way, or not visibly filled at
// all, is not proven safe.
func goBufferEscaped(f diff.File, l diff.Line) bool {
	if langOf(f.Path) != langGo {
		return false
	}
	m := goHTMLBuffer.FindStringSubmatch(l.Text)
	if m == nil {
		return false
	}
	name := m[1] + m[2] + m[3]
	body := goFunctionLines(f.Lines, l.Line)
	writes := 0
	word := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
	use := goBufferUse(name)
	consts := map[string]bool{}
	for _, bl := range body {
		if m := goLocalStrConst.FindStringSubmatch(bl.Text); m != nil {
			consts[m[1]] = true
		}
	}
	for _, bl := range body {
		if IsCommentLine(bl.Text) || !word.MatchString(bl.Text) {
			continue
		}
		if args, ok := use.write(bl.Text); ok {
			writes++
			if !allEscapedArgs(args, consts) {
				return false
			}
			continue
		}
		if !use.benign.MatchString(bl.Text) {
			return false
		}
	}
	return writes > 0
}

// goFunctionLines returns the visible lines of the function holding line n.
func goFunctionLines(lines []diff.Line, n int) []diff.Line {
	at := -1
	for i, l := range lines {
		if l.Line == n {
			at = i
			break
		}
	}
	if at < 0 {
		return nil
	}
	start, end := at, at
	for start > 0 && lines[start-1].Line == lines[start].Line-1 && !goFuncStart.MatchString(lines[start].Text) {
		start--
	}
	for end+1 < len(lines) && lines[end+1].Line == lines[end].Line+1 && !goFuncStart.MatchString(lines[end+1].Text) {
		end++
	}
	return lines[start : end+1]
}

// bufferUse recognizes the statements that touch one buffer variable.
type bufferUse struct {
	writes []*regexp.Regexp // group 1: the written arguments
	benign *regexp.Regexp   // declaration, reset, or reading the content back
}

func goBufferUse(name string) bufferUse {
	n := regexp.QuoteMeta(name)
	return bufferUse{
		writes: []*regexp.Regexp{
			regexp.MustCompile(`\b` + n + `\.(?:WriteString|WriteByte|WriteRune)\((.*)$`),
			regexp.MustCompile(`\b` + n + `\.Write\(\s*\[\]byte\((.*)$`),
			regexp.MustCompile(`\bfmt\.Fprint(?:f|ln)?\(\s*&?` + n + `\s*,(.*)$`),
			regexp.MustCompile(`\bio\.WriteString\(\s*&?` + n + `\s*,(.*)$`),
		},
		benign: regexp.MustCompile(`\bvar\s+` + n + `\s+(?:bytes\.Buffer|strings\.Builder)\b|\b` + n +
			`\s*:?=\s*(?:new\((?:bytes\.Buffer|strings\.Builder)\)|bytes\.NewBuffer(?:String)?\(\s*(?:nil|""|\[\]byte\{\})\s*\)|&?(?:bytes\.Buffer|strings\.Builder)\{)|\b` + n +
			`\.(?:Reset|Len|Grow)\(|\btemplate\.\w+\(\s*(?:string\(\s*)?` + n + `\b`),
	}
}

func (u bufferUse) write(text string) ([]string, bool) {
	for _, re := range u.writes {
		if m := re.FindStringSubmatch(text); m != nil {
			return splitCallArgs(m[1]), true
		}
	}
	return nil, false
}

func allEscapedArgs(args []string, consts map[string]bool) bool {
	for _, a := range args {
		a = strings.TrimSpace(a)
		if strings.Count(a, ")") > strings.Count(a, "(") {
			a = strings.TrimSpace(strings.TrimSuffix(a, ")"))
		}
		if consts[a] || isRuneLiteral(a) {
			continue
		}
		if !isPureLiteral(a) && !coveredByCall(a, goWriter.escape) {
			return false
		}
	}
	return true
}

// isRuneLiteral reports a Go rune literal such as '+' or '\n'.
func isRuneLiteral(a string) bool {
	return len(a) >= 3 && a[0] == '\'' && a[len(a)-1] == '\'' && len(a) <= 6
}
