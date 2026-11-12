package triage

import (
	"fmt"
	"sort"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// Applied records one verdict that changed a report, for the human to audit.
type Applied struct {
	ID     string
	Verb   Verb
	Reason string
	Before finding.Finding
	After  finding.Finding // equals Before for a drop
}

// immune findings can never be dropped or lowered: an injected model must not
// be able to clear them. Untrusted-diff findings are the prompt-injection
// signal itself. Secrets are always CRITICAL and BLOCK; the model reads the
// secret in the hunk, and a placeholder is a rule fix, not a model call.
func immune(c finding.Class) bool { return c == finding.UntrustedDiff || c == finding.Secret }

// Apply returns the findings left after the verdicts and the record of what
// changed. It never adds a finding and never raises a severity. Any verdict
// that names an unknown or ambiguous id, targets an immune class, or does not
// strictly lower the severity, or repeats an id, fails the whole call and returns no findings, so
// a bad verdicts file can never pass silently.
func Apply(findings []finding.Finding, f File) ([]finding.Finding, []Applied, error) {
	byID := map[string][]int{}
	for i, fd := range findings {
		id := finding.ID(fd)
		byID[id] = append(byID[id], i)
	}
	drop := map[int]bool{}
	lowered := map[int]finding.Finding{}
	applied := []Applied{} // non-nil: an empty verdicts file still marks the run as triaged
	seen := map[string]bool{}
	for _, v := range f.Verdicts {
		if err := v.validate(); err != nil {
			return nil, nil, fmt.Errorf("invalid verdict: %w", err)
		}
		if seen[v.ID] {
			return nil, nil, fmt.Errorf("verdict %s: duplicate, one verdict per finding", v.ID)
		}
		seen[v.ID] = true
		idx := byID[v.ID]
		switch {
		case len(idx) == 0:
			return nil, nil, fmt.Errorf("verdict %s: no such finding (a verdict cannot add one)", v.ID)
		case len(idx) > 1:
			return nil, nil, fmt.Errorf("verdict %s: ambiguous id, it matches %d findings", v.ID, len(idx))
		}
		before := findings[idx[0]]
		if immune(before.Class) {
			return nil, nil, fmt.Errorf("verdict %s: %s findings cannot be dropped or downgraded", v.ID, before.Class)
		}
		after := before
		if v.Action == Drop {
			drop[idx[0]] = true
		} else {
			if v.Severity.Rank() <= before.Severity.Rank() {
				return nil, nil, fmt.Errorf("verdict %s: %s does not lower %s (a verdict cannot raise severity)", v.ID, v.Severity, before.Severity)
			}
			after.Severity = v.Severity
			after.Action = finding.ActionFor(v.Severity)
			lowered[idx[0]] = after
		}
		applied = append(applied, Applied{ID: v.ID, Verb: v.Action, Reason: v.Reason, Before: before, After: after})
	}
	out := make([]finding.Finding, 0, len(findings))
	for i, fd := range findings {
		switch {
		case drop[i]:
		case lowered[i].Path != "":
			out = append(out, lowered[i])
		default:
			out = append(out, fd)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return less(out[i], out[j]) })
	return out, applied, nil
}

// less is the engine's report order: severity, then path, then line.
func less(a, b finding.Finding) bool {
	if ra, rb := a.Severity.Rank(), b.Severity.Rank(); ra != rb {
		return ra < rb
	}
	if a.Path != b.Path {
		return a.Path < b.Path
	}
	return a.Line < b.Line
}
