package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	// sqlSafeHelper opens a call that yields table-prefix or placeholder text
	// and never request data: Common::prefixTable(X), getSqlStringFieldsArray($ids),
	// $db->placeholders($ids). Group 1 is the helper name.
	sqlSafeHelper = regexp.MustCompile(`(?:[\w\\]+::|\$?\w+->)?(prefixTable|getSqlStringFieldsArray|\w*[Pp]laceholders?\w*)\s*\(`)
	// placeholderGen builds a list of "?" markers: implode(',', array_fill(0, n, '?')),
	// str_repeat('?,', n), strings.Repeat("?,", n), ",".join("?" * n).
	placeholderGen = regexp.MustCompile(`\b(?:implode|join|array_fill|str_repeat|Repeat|repeat|Join)\b`)
	placeholderLit = regexp.MustCompile(`['"][?,\s]*\?[?,\s]*['"]`)
)

// stripSafeSQLHelpers replaces each table-prefix or placeholder helper call
// with an empty string literal, so the call is not read as dynamic SQL text.
// prefixTable stays when its argument is not a constant: a table name taken
// from a variable is not proven safe.
func stripSafeSQLHelpers(s string) string {
	from := 0
	for from < len(s) {
		loc := sqlSafeHelper.FindStringSubmatchIndex(s[from:])
		if loc == nil {
			return s
		}
		start, nameEnd, open := from+loc[0], from+loc[3], from+loc[1]-1
		end := callEnd(s, open)
		if strings.HasSuffix(s[start:nameEnd], "prefixTable") &&
			len(scanExpr(strings.Join(callArgs(s, open), ","), true).nonLiteralOperands()) > 0 {
			from = open + 1
			continue
		}
		s = s[:start] + `""` + s[end:]
		from = start + 2
	}
	return s
}

// addPlaceholderVars marks each variable assigned a list of "?" markers as
// safe to interpolate into SQL: it carries no request data.
func addPlaceholderVars(safe map[string]bool, lines []diff.Line) {
	for _, l := range lines {
		if IsCommentLine(l.Text) {
			continue
		}
		a, ok := parseAssignment(l.Text)
		if !ok || len(a.names) != 1 || a.appends() {
			continue
		}
		if sqlSafeHelper.MatchString(a.rhs) && stripSafeSQLHelpers(a.rhs) != a.rhs ||
			placeholderGen.MatchString(a.rhs) && placeholderLit.MatchString(a.rhs) {
			safe[a.names[0]] = true
		}
	}
}
