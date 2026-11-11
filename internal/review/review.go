// Package review is the deterministic engine: unified diff in, findings out.
package review

import (
	"fmt"
	"sort"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/rules"
)

// Run parses the diff, applies every rule to every file, dedupes by
// path:line:cwe (first wins) and sorts by severity, path, then line.
// Ordering by path is a byte compare.
func Run(text string, rs []rules.Rule) []finding.Finding {
	var all []finding.Finding
	for _, file := range diff.Parse(text) {
		for _, rule := range rs {
			all = append(all, rule(file)...)
		}
	}
	out := dedupe(all)
	sort.SliceStable(out, func(i, j int) bool { return less(out[i], out[j]) })
	return out
}

func less(a, b finding.Finding) bool {
	if ra, rb := a.Severity.Rank(), b.Severity.Rank(); ra != rb {
		return ra < rb
	}
	if a.Path != b.Path {
		return a.Path < b.Path
	}
	return a.Line < b.Line
}

func dedupe(in []finding.Finding) []finding.Finding {
	seen := make(map[string]bool, len(in))
	out := make([]finding.Finding, 0, len(in))
	for _, f := range in {
		key := fmt.Sprintf("%s:%d:%s", f.Path, f.Line, f.CWE)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, f)
	}
	return out
}
