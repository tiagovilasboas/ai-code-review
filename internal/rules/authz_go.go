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
	// A file name is not an object id: names cleaned by Base and file-system
	// calls belong to path traversal, not to AuthZ.
	ginBaseAssign  = regexp.MustCompile(`\b(?:filepath|path)\.Base\(`)
	ginFileOp      = regexp.MustCompile(`\b(?:os|ioutil|filepath|path|io|fs)\.\w+\(|\.(?:File|FileFS|Attachment|Inline|FileAttachment)\(`)
	ginParamCall   = regexp.MustCompile(`\.(?:Param|ByName|PathValue|URLParam)\(\s*(?:\w+\s*,\s*)?"(\w+)"`)
	ginInlineParam = regexp.MustCompile(`\.(?:Param|ByName|PathValue|URLParam)\(`)
	// Repository, ORM or service call that reads or changes one object.
	ginSink = regexp.MustCompile(`\.(?:Get|Find|First|Take|Last|Update|Save|Delete|Remove|Load|Fetch|Read|Patch|Where)\w*\(|\.\w+By(?:ID|Id|UUID|Uuid)\(`)
	// The authenticated user taken from the gin context or token claims, or an
	// ownership/tenant field (also inside a SQL string).
	// restBuilder is a REST client request builder (client-go: c.Get().Namespace(ns).Name(n).Do(ctx)):
	// an HTTP call made on the caller's behalf, not a read of a stored object.
	restBuilder = regexp.MustCompile(`\.(?:Get|Post|Put|Patch|Delete)\(\s*\)\s*(?:\.\s*(?:Namespace|Name|Resource|SubResource|Suffix|Prefix|AbsPath|Param|VersionedParams|Body|Timeout|Do|DoRaw|Stream|Watch)\s*\(|\.\s*$)`)
	// forwardsToken: the handler builds an upstream client with the caller's own
	// bearer token, so the upstream API authorizes each request.
	forwardsToken = regexp.MustCompile(`\bBearerToken\b|\bSetBearerToken\s*\(|\.Set\(\s*"Authorization"`)
	// convertsID is a call that parses a route id into its typed form.
	convertsID   = regexp.MustCompile(`\bstrconv\.\w+\(|\buuid\.\w+\(|\b(?:Parse\w*|Atoi|MustParse|ObjectIDFromHex)\(|\b(?:int|int64|uint|uint64)\(`)
	ginPrincipal = regexp.MustCompile(`(?i)\.(?:MustGet|GetString|GetInt\d*|GetUint\d*)\(|\bc\.Get\(|user_?id|owner_?id|tenant_?id|currentuser|claims|principal|authuser|session|Context\(\)\.Value\(|\b(?:user|usr|u|me|acct|account|actor|caller|member)\.(?:ID|Id|UserID)\b`)
)

// ginUnchecked returns the first id-keyed lookup line of every gin handler
// that never consults the authenticated user.
func ginUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	seen := map[int]bool{}
	for i, l := range lines {
		if IsCommentLine(l.Text) || !ginParamCall.MatchString(l.Text) { // the route param may be a context line
			continue
		}
		scope := enclosingBy(lines, i, func(t string) bool { return goFuncOpener.MatchString(t) })
		if seen[scope[0].Line] {
			continue
		}
		seen[scope[0].Line] = true
		ids := ginCallerIDs(scope)
		sink, ok := firstSink(withoutRESTChains(withoutFileOps(scope)), ginSink, ids)
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
		if ginBaseAssign.MatchString(l.Text) {
			continue // the param is a file name cleaned with Base
		}
		if m := ginParamAssign.FindStringSubmatch(l.Text); m != nil && m[1] != "_" && m[2] != "slug" {
			ids = append(ids, wordRegexp(m[1]))
		} else if m := ginParamCall.FindStringSubmatch(l.Text); m != nil && m[1] != "slug" && !ginParamAssign.MatchString(l.Text) {
			ids = append(ids, ginInlineParam)
		}
	}
	return append(ids, convertedIDs(scope, ids)...)
}

// convertedIDs matches the variables a tracked id is parsed into
// (id, err := strconv.ParseInt(idStr, 10, 64)).
func convertedIDs(scope []diff.Line, ids []*regexp.Regexp) []*regexp.Regexp {
	var out []*regexp.Regexp
	all := append([]*regexp.Regexp{}, ids...)
	for _, l := range scope {
		a, ok := parseAssignment(stripStrings(l.Text))
		if !ok || IsCommentLine(l.Text) || !convertsID.MatchString(a.rhs) || a.names[0] == "_" || ginParamCall.MatchString(l.Text) {
			continue
		}
		for _, id := range all {
			if id != ginInlineParam && id.MatchString(a.rhs) {
				re := wordRegexp(a.names[0])
				out, all = append(out, re), append(all, re)
				break
			}
		}
	}
	return out
}

func ginGuarded(scope []diff.Line) bool {
	var rest []diff.Line
	for _, l := range scope {
		if !ginParamCall.MatchString(l.Text) {
			rest = append(rest, l) // a "userID" route param name is not a principal
		}
	}
	for _, l := range rest {
		if !IsCommentLine(l.Text) && forwardsToken.MatchString(l.Text) {
			return true
		}
	}
	return hasGuard(rest, ginPrincipal)
}

// withoutRESTChains drops REST client builder lines from the sink search.
func withoutRESTChains(scope []diff.Line) []diff.Line {
	var rest []diff.Line
	for _, l := range scope {
		if !restBuilder.MatchString(stripStrings(l.Text)) {
			rest = append(rest, l)
		}
	}
	return rest
}

// withoutFileOps drops file-system call lines from the sink search; the
// guard search keeps the whole handler.
func withoutFileOps(scope []diff.Line) []diff.Line {
	var rest []diff.Line
	for _, l := range scope {
		if !ginFileOp.MatchString(l.Text) {
			rest = append(rest, l)
		}
	}
	return rest
}
