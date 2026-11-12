package report

import (
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/triage"
)

type sarifSuppression struct {
	Kind          string `json:"kind"`
	Status        string `json:"status"`
	Justification string `json:"justification"`
}

type sarifResultProps struct {
	Triage sarifTriageProps `json:"triage"`
}

type sarifTriageProps struct {
	Verdict      triage.Verb      `json:"verdict"`
	Reason       string           `json:"reason"`
	FromSeverity finding.Severity `json:"fromSeverity"`
	ToSeverity   finding.Severity `json:"toSeverity,omitempty"`
}

// withDropped returns the findings plus the dropped originals, so a dropped
// finding's class still has a rule to point at.
func withDropped(findings []finding.Finding, applied []triage.Applied) []finding.Finding {
	out := append([]finding.Finding(nil), findings...)
	for _, a := range applied {
		if a.Verb == triage.Drop {
			out = append(out, a.Before)
		}
	}
	return out
}

// applySarifTriage marks a downgraded result with its verdict.
func applySarifTriage(r *sarifResult, f finding.Finding, applied []triage.Applied) {
	for _, a := range applied {
		if a.Verb == triage.Downgrade && a.ID == finding.ID(f) {
			r.Properties = &sarifResultProps{Triage: sarifTriageProps{
				Verdict: a.Verb, Reason: a.Reason, FromSeverity: a.Before.Severity, ToSeverity: a.After.Severity,
			}}
		}
	}
}

// droppedResults keeps each dropped finding visible as a suppressed result.
func droppedResults(applied []triage.Applied, index map[finding.Class]int) []sarifResult {
	var out []sarifResult
	for _, a := range applied {
		if a.Verb != triage.Drop {
			continue
		}
		r := sarifResultFor(a.Before, index[a.Before.Class])
		r.Suppressions = []sarifSuppression{{Kind: "external", Status: "accepted", Justification: a.Reason}}
		r.Properties = &sarifResultProps{Triage: sarifTriageProps{
			Verdict: a.Verb, Reason: a.Reason, FromSeverity: a.Before.Severity,
		}}
		out = append(out, r)
	}
	return out
}
