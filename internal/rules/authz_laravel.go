package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	// Controller action or route closure: group 1 is the parameter list.
	laravelAction  = regexp.MustCompile(`^\s*(?:public\s+)?function\s+(?:\w+)\s*\(([^)]*)\)`)
	laravelClosure = regexp.MustCompile(`\bRoute::\w+\(.*\bfunction\s*\(([^)]*)\)`)
	laravelIDParam = regexp.MustCompile(`\$(\w*[iI]d|id)\b`)
	// $oid = $request->input('id'), ->route('order_id'); and the property form $request->id.
	laravelReqAssign = regexp.MustCompile(`\$(\w+)\s*=\s*\$request->(?:input|route|query|get)\(\s*['"]\w*(?i:id)['"]`)
	laravelReqProp   = regexp.MustCompile(`\$request->(?:id|\w+_id|\w+Id)\b`)
	laravelSink      = regexp.MustCompile(`\b[A-Z]\w*::(?:find|findOrFail|findOrNew|firstOrFail|destroy|where\w*|findMany)\(|->(?:find|findOrFail|firstOrFail|update|delete|destroy)\(`)
	// $request->user(), auth(), policy or gate calls, 403 aborts, owner fields,
	// public-by-design flags.
	laravelPrincipal = regexp.MustCompile(`\$request->user\(|\bauth\(|\bAuth::(?:user|id|check)|\$this->authorize\w*\(|\bGate::\w+\(|->(?:can|cannot)\(|\babort_(?:if|unless)\(|\babort\(\s*40[13]\b|\$(?:user|currentUser)->id\b|\b(?:user_id|owner_id|tenant_id|team_id)\b|\bis_(?:published|public)\b|\bAuthorizationException\b`)
)

// laravelUnchecked returns the first id-keyed lookup of every Laravel action
// that never consults the authenticated user, a policy or a gate.
func laravelUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for i, l := range lines {
		if l.Kind != diff.Added || IsCommentLine(l.Text) {
			continue
		}
		m := laravelAction.FindStringSubmatch(l.Text)
		if m == nil {
			m = laravelClosure.FindStringSubmatch(l.Text)
		}
		if m == nil {
			continue
		}
		block := braceBlock(lines, i, 3)
		if block == nil {
			continue
		}
		var ids []*regexp.Regexp
		for _, p := range laravelIDParam.FindAllStringSubmatch(m[1], -1) {
			ids = append(ids, regexp.MustCompile(`\$`+p[1]+`\b`))
		}
		ids = append(ids, laravelReqProp)
		for _, bl := range block {
			if a := laravelReqAssign.FindStringSubmatch(bl.Text); a != nil {
				ids = append(ids, regexp.MustCompile(`\$`+a[1]+`\b`))
			}
		}
		if sink, ok := firstSink(block[1:], laravelSink, ids); ok && !hasGuard(block[1:], laravelPrincipal) {
			out = append(out, sink)
		}
	}
	return out
}
