package rules

import (
	"path"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// proseExts are documents nothing executes; .mdx is left out because it holds JSX.
var proseExts = map[string]bool{".md": true, ".markdown": true, ".txt": true, ".rst": true, ".adoc": true}

// SkipProse wraps a code-sink rule so it stays silent on prose documents: a
// sink quoted in a README or a report is text about code, not code.
func SkipProse(r Rule) Rule {
	return func(f diff.File) []finding.Finding {
		if proseExts[strings.ToLower(path.Ext(f.Path))] {
			return nil
		}
		return r(f)
	}
}
