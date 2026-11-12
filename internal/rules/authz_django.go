package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// Django function and class views: def view(request, pk) with an ORM lookup.
var (
	djangoView = regexp.MustCompile(`^\s*(?:async\s+)?def\s+\w+\s*\(\s*(?:self\s*,\s*)?request\s*,([^)]*)\)`)
	djangoID   = regexp.MustCompile(`(?:^|[,\s])(pk|id|\w+_(?:id|pk))\s*(?:[:=,)]|$)`)
)

// djangoUnchecked returns the first id-keyed lookup of every Django view that
// takes an id from the URL and never consults request.user or an owner field.
func djangoUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for i, l := range lines {
		m := djangoView.FindStringSubmatch(l.Text)
		if l.Kind != diff.Added || m == nil {
			continue
		}
		var ids []*regexp.Regexp
		for _, id := range djangoID.FindAllStringSubmatch(m[1], -1) {
			ids = append(ids, wordRegexp(id[1]))
		}
		view := pyBlock(lines, i, i)
		if len(ids) == 0 {
			continue
		}
		if sink, ok := firstSink(view[1:], flaskSink, ids); ok && !hasGuard(view[1:], flaskPrincipal) {
			out = append(out, sink)
		}
	}
	return out
}
