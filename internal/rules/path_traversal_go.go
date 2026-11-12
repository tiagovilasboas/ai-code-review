package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	goIsLocal   = regexp.MustCompile(`\bfilepath\.IsLocal\s*\(`)
	goRel       = regexp.MustCompile(`\bfilepath\.Rel\s*\(`)
	goDotDot    = regexp.MustCompile(`"\.\."`)
	goJoin      = regexp.MustCompile(`\bfilepath\.Join\s*\(`)
	goHasPrefix = regexp.MustCompile(`\bstrings\.HasPrefix\s*\(`)
	// goRootClean: Clean("/"+name) resolves every ".." against the root first.
	goRootClean = regexp.MustCompile(`\b(?:filepath|path)\.Clean\(\s*(?:"/"|string\(\s*(?:filepath\.Separator|os\.PathSeparator)\s*\))\s*\+`)
)

// goContained reports a Go containment check in the hunk: filepath.IsLocal,
// filepath.Rel compared with "..", or a filepath.Join result (already cleaned)
// tested with strings.HasPrefix, or Clean("/"+name) before the join.
func goContained(hunk []diff.Line) bool {
	var isLocal, rel, dotDot, join, prefix, rootClean bool
	for _, l := range hunk {
		if IsCommentLine(l.Text) {
			continue
		}
		isLocal = isLocal || goIsLocal.MatchString(l.Text)
		rel = rel || goRel.MatchString(l.Text)
		dotDot = dotDot || goDotDot.MatchString(l.Text)
		join = join || goJoin.MatchString(l.Text)
		prefix = prefix || goHasPrefix.MatchString(l.Text)
		rootClean = rootClean || goRootClean.MatchString(l.Text)
	}
	return isLocal || rootClean || (rel && dotDot) || (join && prefix)
}
