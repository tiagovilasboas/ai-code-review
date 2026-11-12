package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// frameworkAuthZ flags handlers of Go (gin/chi), NestJS, Flask/FastAPI/Django, Spring and Laravel that look up or
// change an object by a caller-supplied id without any ownership evidence in
// the same handler. Each handler yields at most one finding, on its first
// lookup or mutation line.
func frameworkAuthZ(f diff.File) []finding.Finding {
	if isTestPath(f.Path) {
		return nil
	}
	var hits []diff.Line
	switch {
	case strings.HasSuffix(f.Path, ".go"):
		hits = ginUnchecked(f.Lines)
	case strings.HasSuffix(f.Path, ".ts"):
		hits = nestUnchecked(f.Lines)
	case strings.HasSuffix(f.Path, ".py"):
		hits = append(flaskUnchecked(f.Lines), djangoUnchecked(f.Lines)...)
	case strings.HasSuffix(f.Path, ".java"):
		hits = springUnchecked(f.Lines)
	case strings.HasSuffix(f.Path, ".php"):
		hits = laravelUnchecked(f.Lines)
	}
	var out []finding.Finding
	for _, l := range hits {
		out = add(out, finding.Input{
			Title:    "Missing object-level authorization",
			Class:    finding.AuthZ,
			Path:     f.Path,
			Line:     l.Line,
			CWE:      "CWE-639",
			OWASP:    finding.LabelsFor(finding.AuthZ),
			Severity: finding.High,
		})
	}
	return out
}

// mergeByLine appends extra findings whose line is not already reported.
func mergeByLine(base, extra []finding.Finding) []finding.Finding {
	seen := map[int]bool{}
	for _, b := range base {
		seen[b.Line] = true
	}
	for _, e := range extra {
		if !seen[e.Line] {
			base = append(base, e)
		}
	}
	return base
}

// braceBlock returns lines from start until the braces opened on or after it
// balance, within one contiguous hunk. It is nil when no brace opens in the
// first openWithin lines.
func braceBlock(lines []diff.Line, start, openWithin int) []diff.Line {
	depth, opened := 0, false
	for i := start; i < len(lines); i++ {
		if i > start && lines[i].Line != lines[i-1].Line+1 {
			break
		}
		if !IsCommentLine(lines[i].Text) {
			depth += braceNet(lines[i].Text)
			opened = opened || depth > 0
		}
		if opened && depth <= 0 {
			return lines[start : i+1]
		}
		if !opened && i-start >= openWithin {
			return nil
		}
	}
	return nil
}

// firstSink returns the first added, non-comment line in scope that matches
// sink and mentions one of the caller ids.
func firstSink(scope []diff.Line, sink *regexp.Regexp, ids []*regexp.Regexp) (diff.Line, bool) {
	for _, l := range scope {
		code := stripStrings(l.Text)
		if l.Kind != diff.Added || IsCommentLine(l.Text) || !sink.MatchString(code) {
			continue
		}
		for _, id := range ids {
			if id.MatchString(code) {
				return l, true
			}
		}
	}
	return diff.Line{}, false
}

// hasGuard reports ownership or permission evidence anywhere in scope, using
// the framework principal pattern plus the shared admin and policy guards.
func hasGuard(scope []diff.Line, principal *regexp.Regexp) bool {
	for _, l := range scope {
		if IsCommentLine(l.Text) {
			continue
		}
		if principal.MatchString(l.Text) || authzInScope.MatchString(stripStrings(l.Text)) || authzAdminGuard.MatchString(l.Text) {
			return true
		}
	}
	return false
}

func wordRegexp(name string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
}
