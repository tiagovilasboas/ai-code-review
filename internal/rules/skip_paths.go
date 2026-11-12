package rules

import (
	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// SkipGenerated wraps a code-sink rule so it stays silent on build artifacts
// (see isGeneratedPath) and on added lines past maxSourceLine. It is meant for
// content-sink rules only: secrets and untrusted-diff keep reading these files,
// because a leaked key or an injected instruction is as real in a bundle.
func SkipGenerated(r Rule) Rule {
	return func(f diff.File) []finding.Finding {
		if isGeneratedPath(f.Path) {
			return nil
		}
		return r(withoutLongLines(f))
	}
}

// SkipTests wraps a code-sink rule so it stays silent on test files: a sink
// there mounts a fixture and is never reachable from a request.
func SkipTests(r Rule) Rule {
	return func(f diff.File) []finding.Finding {
		if isTestPath(f.Path) {
			return nil
		}
		return r(f)
	}
}
