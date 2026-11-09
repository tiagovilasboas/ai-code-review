// Package rules holds the deterministic fail-closed checks, one file per sink
// family. Each rule is a Rule and is registered by the review engine.
//
// Contract (each file exports exactly one of these):
//
//	XSS, SSRF, SupplyChain, AuthZ, Secrets, UntrustedDiff,
//	SQLInjection, CmdInjection, PathTraversal
package rules

import (
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// Rule inspects one changed file and returns evidence-backed findings.
type Rule func(f diff.File) []finding.Finding

// IsCommentLine reports whether a source line is a comment (comments are data, not sinks).
func IsCommentLine(text string) bool {
	t := strings.TrimSpace(text)
	for _, p := range []string{"//", "#", "*", "/*", "<!--"} {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

// IsStringLiteral reports whether expr is a constant string literal.
func IsStringLiteral(expr string) bool {
	t := strings.TrimSpace(expr)
	if len(t) < 2 {
		return false
	}
	first, last := t[0], t[len(t)-1]
	switch {
	case first == '"' && last == '"', first == '\'' && last == '\'':
		return true
	case first == '`' && last == '`':
		return !strings.Contains(t, "${")
	}
	return false
}

// add appends in to out only when finding.New accepts it.
func add(out []finding.Finding, in finding.Input) []finding.Finding {
	if f, ok := finding.New(in); ok {
		return append(out, f)
	}
	return out
}
