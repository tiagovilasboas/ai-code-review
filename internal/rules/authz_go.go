package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	goFuncOpener = regexp.MustCompile(`\bfunc\b.*\{\s*$`)
	// (also chi.URLParam(r, "id") and r.PathValue("id"))
	// id := c.Param("id"), id, err := strconv.Atoi(c.Param("id")); group 1 is the
	// assigned name, group 2 the route parameter.
	ginParamAssign = regexp.MustCompile(`^\s*(\w+)(?:\s*,\s*\w+)?\s*:?=.*\.(?:Param|ByName|PathValue|URLParam)\(\s*(?:\w+\s*,\s*)?"(\w+)"`)
	ginParamCall   = regexp.MustCompile(`\.(?:Param|ByName|PathValue|URLParam)\(\s*(?:\w+\s*,\s*)?"(\w+)"`)
	ginInlineParam = regexp.MustCompile(`\.(?:Param|ByName|PathValue|URLParam)\(`)
	// Repository, ORM or service call that reads or changes one object.
	ginSink = regexp.MustCompile(`\.(?:Get|Find|First|Take|Last|Update|Save|Delete|Remove|Load|Fetch|Read|Patch|Where)\w*\(`)
	// The authenticated user taken from the gin context or token claims, or an
	// ownership/tenant field (also inside a SQL string).
	ginPrincipal = regexp.MustCompile(`(?i)\.(?:MustGet|GetString|GetInt\d*|GetUint\d*)\(|\bc\.Get\(|user_?id|owner_?id|tenant_?id|currentuser|claims|principal|authuser|session|Context\(\)\.Value\(`)
)

// ginUnchecked returns the first id-keyed lookup line of every gin handler
// that never consults the authenticated user.
func ginUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	seen := map[int]bool{}
	for i, l := range lines {
		if l.Kind != diff.Added || IsCommentLine(l.Text) || !ginParamCall.MatchString(l.Text) {
			continue
		}
		scope := enclosingBy(lines, i, func(t string) bool { return goFuncOpener.MatchString(t) })
		if seen[scope[0].Line] {
			continue
		}
		seen[scope[0].Line] = true
		ids := ginCallerIDs(scope)
		sink, ok := firstSink(scope, ginSink, ids)
		if ok && !ginGuarded(scope) {
			out = append(out, sink)
		}
	}
	return out
}

// ginCallerIDs matches the variables bound to non-slug route params, plus an
// inline Param call.
func ginCallerIDs(scope []diff.Line) []*regexp.Regexp {
	var ids []*regexp.Regexp
	for _, l := range scope {
		if m := ginParamAssign.FindStringSubmatch(l.Text); m != nil && m[1] != "_" && m[2] != "slug" {
			ids = append(ids, wordRegexp(m[1]))
		} else if m := ginParamCall.FindStringSubmatch(l.Text); m != nil && m[1] != "slug" && !ginParamAssign.MatchString(l.Text) {
			ids = append(ids, ginInlineParam)
		}
	}
	return ids
}

func ginGuarded(scope []diff.Line) bool {
	var rest []diff.Line
	for _, l := range scope {
		if !ginParamCall.MatchString(l.Text) {
			rest = append(rest, l) // a "userID" route param name is not a principal
		}
	}
	return hasGuard(rest, ginPrincipal)
}
