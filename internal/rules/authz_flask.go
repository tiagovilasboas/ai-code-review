package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	flaskRoute  = regexp.MustCompile(`^\s*@\w+(?:\.\w+)*\.(?:route|get|post|put|patch|delete)\s*\(\s*['"]([^'"]*)['"]`)
	flaskURLVar = regexp.MustCompile(`<(?:\w+(?:\([^)]*\))?:)?(\w+)>|\{(\w+)(?::[^}]*)?\}`) // Flask <id> or FastAPI {id}
	flaskDef    = regexp.MustCompile(`^\s*(?:async\s+)?def\s+\w+\s*\(`)
	flaskSink   = regexp.MustCompile(`\.query\.(?:get|get_or_404|filter\w*|first_or_404)\b|\bget_or_404\(|\bfirst_or_404\(|\.filter_by\(|\bsession\.(?:get|query)\(|\.objects\.(?:get|filter)\(|\b[A-Z]\w*\.(?:get|get_by_id|find|find_one|find_by_id)\(|\bget_object_or_404\(|\bdb\.(?:get|query|scalar|scalars|execute)\(|\b(?:crud|repo|repository)\.\w*(?:get|read|update|delete|find)\w*\(`)
	// current_user and friends, ownership fields, explicit 401/403, role decorators.
	flaskPrincipal = regexp.MustCompile(`\b(?:current_user|current_identity|get_jwt_identity|g\.user|g\.current_user|request\.user|owner_id|user_id|tenant_id)\b|\babort\(\s*40[13]\b|\bForbidden\b|@(?:admin_required|roles_required|roles_accepted|permission_required|requires_\w+)\b|\bis_?(?:public|published)\s*=\s*True|\bDepends\(\s*\w*(?:current|auth|owner|permission)\w*`)
)

// flaskUnchecked returns the first id-keyed lookup line of every Flask view
// whose URL carries an id and that never consults the current user.
func flaskUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for i, l := range lines {
		m := flaskRoute.FindStringSubmatch(l.Text)
		if l.Kind != diff.Added || m == nil {
			continue
		}
		ids := flaskURLIDs(m[1])
		view := flaskView(lines, i)
		if len(ids) == 0 || view == nil {
			continue
		}
		sink, ok := firstSink(view, flaskSink, ids)
		if ok && !hasGuard(view, flaskPrincipal) {
			out = append(out, sink)
		}
	}
	return out
}

// flaskURLIDs matches the non-slug converters of a route path.
func flaskURLIDs(path string) []*regexp.Regexp {
	var ids []*regexp.Regexp
	for _, m := range flaskURLVar.FindAllStringSubmatch(path, -1) {
		if name := m[1] + m[2]; name != "slug" {
			ids = append(ids, wordRegexp(name))
		}
	}
	return ids
}

// flaskView returns the route decorators, the def line and the indented body
// that follows lines[index], within one hunk.
func flaskView(lines []diff.Line, index int) []diff.Line {
	def := -1
	for i := index + 1; i < len(lines) && i <= index+10 && lines[i].Line == lines[i-1].Line+1; i++ {
		if flaskDef.MatchString(lines[i].Text) {
			def = i
			break
		}
	}
	if def < 0 {
		return nil
	}
	return pyBlock(lines, index, def)
}

// pyBlock returns lines[from:] through the end of the indented body of the
// def at lines[def], within one hunk.
func pyBlock(lines []diff.Line, from, def int) []diff.Line {
	indent := leadingSpaces(lines[def].Text)
	end := def
	for i := def + 1; i < len(lines) && lines[i].Line == lines[i-1].Line+1; i++ {
		if strings.TrimSpace(lines[i].Text) != "" && leadingSpaces(lines[i].Text) <= indent {
			break
		}
		end = i
	}
	return lines[from : end+1]
}

func leadingSpaces(text string) int {
	return len(text) - len(strings.TrimLeft(text, " \t"))
}
