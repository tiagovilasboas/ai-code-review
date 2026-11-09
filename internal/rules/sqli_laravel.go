package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// laravelRawSink is a Laravel call whose first argument is raw SQL (or a raw
// SQL fragment). Later arguments are bindings, so they are never read.
var laravelRawSink = regexp.MustCompile(`\bDB::(?:select|insert|update|delete|statement|unprepared|raw)\s*\(|` +
	`(?:->|::)(?:whereRaw|orWhereRaw|havingRaw|orHavingRaw|selectRaw|orderByRaw|groupByRaw|fromRaw|joinRaw)\s*\(`)

// laravelRawSQL inspects a Laravel raw-SQL sink on lines[i]. matched reports
// that the line has one; dynamic that its SQL is built from non-literal
// values or is request input; fromRequest that request input reaches it.
func laravelRawSQL(lines []diff.Line, i int, tainted, built, safe map[string]bool) (matched, dynamic, fromRequest bool) {
	loc := laravelRawSink.FindStringIndex(lines[i].Text)
	if loc == nil {
		return false, false, false
	}
	args := callArgs(callFrom(lines, i, 0), loc[1]-1)
	if len(args) == 0 || sqlEscaped.MatchString(args[0]) {
		return true, false, false
	}
	e := scanExpr(args[0], true)
	fromRequest = reachesRequest(maskAllowlisted(args[0], safe), tainted, true)
	dynamic = len(dropAllowlisted(e.dynamicOperands(true), safe)) > 0 || fromRequest
	for name, fromReq := range built {
		if mentions(e.code, name) {
			dynamic = true
			fromRequest = fromRequest || fromReq
		}
	}
	return true, dynamic, dynamic && fromRequest
}
