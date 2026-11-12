package report

import (
	"fmt"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/triage"
)

// triageNotice tells the reader why the section exists.
const triageNotice = "Triaged: verdicts from an external model changed this report. A human must review each one."

func formatTriaged(applied []triage.Applied) string {
	if len(applied) == 0 {
		return ""
	}
	lines := []string{"", triageNotice}
	for _, a := range applied {
		b := a.Before
		head := fmt.Sprintf("- %s %s at %s:%d (%s, %s %s)", a.Verb, b.Class.Label(), b.Path, b.Line, a.ID, b.Severity, b.Action)
		if a.Verb == triage.Downgrade {
			head += fmt.Sprintf(" -> %s %s", a.After.Severity, a.After.Action)
		}
		lines = append(lines, head, "  Reason: "+a.Reason)
	}
	return strings.Join(lines, "\n") + "\n"
}

type jsonTriage struct {
	Applied []jsonApplied `json:"applied"`
}

type jsonApplied struct {
	ID           string           `json:"id"`
	Verdict      triage.Verb      `json:"verdict"`
	Reason       string           `json:"reason"`
	Class        string           `json:"class"`
	FindingClass finding.Class    `json:"findingClass"`
	Path         string           `json:"path"`
	Line         int              `json:"line"`
	FromSeverity finding.Severity `json:"fromSeverity"`
	FromAction   finding.Action   `json:"fromAction"`
	ToSeverity   finding.Severity `json:"toSeverity,omitempty"`
	ToAction     finding.Action   `json:"toAction,omitempty"`
}

// jsonTriageFor is nil without --triage so plain reports stay unchanged.
func jsonTriageFor(applied []triage.Applied) *jsonTriage {
	if applied == nil {
		return nil
	}
	t := &jsonTriage{Applied: make([]jsonApplied, 0, len(applied))}
	for _, a := range applied {
		b := a.Before
		j := jsonApplied{
			ID: a.ID, Verdict: a.Verb, Reason: a.Reason, Class: b.Class.Label(),
			FindingClass: b.Class, Path: b.Path, Line: b.Line,
			FromSeverity: b.Severity, FromAction: b.Action,
		}
		if a.Verb == triage.Downgrade {
			j.ToSeverity, j.ToAction = a.After.Severity, a.After.Action
		}
		t.Applied = append(t.Applied, j)
	}
	return t
}
