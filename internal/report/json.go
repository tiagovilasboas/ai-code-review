package report

import (
	"encoding/json"
	"fmt"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// ToolName is the tool identity shared by the JSON and SARIF reports.
const ToolName = "ai-code-review"

type jsonReport struct {
	SchemaVersion int           `json:"schemaVersion"`
	Tool          jsonTool      `json:"tool"`
	Decision      jsonDecision  `json:"decision"`
	Findings      []jsonFinding `json:"findings"`
}

type jsonTool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type jsonDecision struct {
	Merge   bool   `json:"merge"`
	Reason  string `json:"reason,omitempty"`
	Blocker string `json:"blocker,omitempty"`
}

type jsonFinding struct {
	Title        string           `json:"title"`
	Class        string           `json:"class"`
	FindingClass finding.Class    `json:"findingClass"`
	Path         string           `json:"path"`
	Line         int              `json:"line"`
	CWE          string           `json:"cwe"`
	OWASP        []string         `json:"owasp"`
	Severity     finding.Severity `json:"severity"`
	Action       finding.Action   `json:"action"`
}

// FormatJSON renders the schemaVersion 1 report, findings in engine order.
func FormatJSON(findings []finding.Finding, version string) (string, error) {
	r := jsonReport{
		SchemaVersion: 1,
		Tool:          jsonTool{Name: ToolName, Version: version},
		Decision:      jsonDecision{Merge: true},
		Findings:      make([]jsonFinding, 0, len(findings)),
	}
	if b, ok := PickMergeBlocker(findings); ok {
		r.Decision = jsonDecision{
			Merge:   false,
			Reason:  fmt.Sprintf("%s at %s:%d (%s)", b.Class.Label(), b.Path, b.Line, asvsOrOfficial(b)),
			Blocker: fmt.Sprintf("%s:%d", b.Path, b.Line),
		}
	}
	for _, f := range findings {
		r.Findings = append(r.Findings, jsonFinding{
			Title: f.Title, Class: f.Class.Label(), FindingClass: f.Class,
			Path: f.Path, Line: f.Line, CWE: f.CWE, OWASP: f.OWASP,
			Severity: f.Severity, Action: f.Action,
		})
	}
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}
