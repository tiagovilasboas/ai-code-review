package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func formatsFinding(t *testing.T, c finding.Class, path string, line int, cwe string, sev finding.Severity) finding.Finding {
	t.Helper()
	f, ok := finding.New(finding.Input{
		Title: "T " + string(c), Class: c, Path: path, Line: line, CWE: cwe,
		OWASP: finding.LabelsFor(c), Severity: sev,
	})
	if !ok {
		t.Fatalf("invalid finding %s", c)
	}
	return f
}

func sample(t *testing.T) []finding.Finding {
	return []finding.Finding{
		formatsFinding(t, finding.AuthZ, "src/o.ts", 36, "CWE-639", finding.High),
		formatsFinding(t, finding.XSS, "src/x.ts", 4, "CWE-79", finding.Medium),
		formatsFinding(t, finding.AuthZ, "src/o.ts", 50, "CWE-639", finding.Critical),
	}
}

func TestJSONBlock(t *testing.T) {
	out, err := FormatJSON(sample(t), "1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		SchemaVersion int
		Tool          struct{ Name, Version string }
		Decision      struct {
			Merge           bool
			Reason, Blocker string
		}
		Findings []map[string]any
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatal(err)
	}
	if r.SchemaVersion != 1 || r.Tool.Name != "ai-code-review" || r.Tool.Version != "1.2.3" {
		t.Fatalf("header: %+v", r)
	}
	if r.Decision.Merge || r.Decision.Blocker != "src/o.ts:36" ||
		r.Decision.Reason != "AuthZ/IDOR at src/o.ts:36 (ASVS-5.0-8.2.2)" {
		t.Fatalf("decision: %+v", r.Decision)
	}
	if len(r.Findings) != 3 || r.Findings[0]["class"] != "AuthZ/IDOR" || r.Findings[0]["findingClass"] != "authz" {
		t.Fatalf("findings: %v", r.Findings)
	}
}

func TestJSONNoFindings(t *testing.T) {
	out, _ := FormatJSON(nil, "dev")
	if !strings.Contains(out, `"merge": true`) || !strings.Contains(out, `"findings": []`) {
		t.Fatalf("got %s", out)
	}
}

func TestSARIFStructure(t *testing.T) {
	out, err := FormatSARIF(sample(t), "dev")
	if err != nil {
		t.Fatal(err)
	}
	var log struct {
		Schema  string `json:"$schema"`
		Version string
		Runs    []struct {
			Tool struct {
				Driver struct {
					Name  string
					Rules []struct {
						ID, Name         string
						ShortDescription struct{ Text string }
						HelpURI          string
						Properties       map[string]any
					}
				}
			}
			Results []struct {
				RuleID    string
				RuleIndex int
				Level     string
				Message   struct{ Text string }
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct{ URI string }
						Region           struct{ StartLine int }
					}
				}
				PartialFingerprints map[string]string
			}
		}
	}
	if err := json.Unmarshal([]byte(out), &log); err != nil {
		t.Fatal(err)
	}
	if log.Schema != "https://json.schemastore.org/sarif-2.1.0.json" || log.Version != "2.1.0" || len(log.Runs) != 1 {
		t.Fatalf("header: %+v", log)
	}
	run := log.Runs[0]
	if run.Tool.Driver.Name != "ai-code-review" || len(run.Tool.Driver.Rules) != 2 {
		t.Fatalf("driver: %+v", run.Tool.Driver)
	}
	r0 := run.Tool.Driver.Rules[0]
	if r0.ID != "authz" || r0.Name != "AuthZ/IDOR" || r0.ShortDescription.Text != "AuthZ/IDOR" ||
		r0.HelpURI != "https://github.com/tiagovilasboas/ai-code-review/blob/main/skills/authz-idor.md" {
		t.Fatalf("rule: %+v", r0)
	}
	if r0.Properties["security-severity"] != "9.5" {
		t.Fatalf("severity: %v", r0.Properties)
	}
	if tags, _ := r0.Properties["tags"].([]any); len(tags) < 3 || tags[0] != "A01:2021" {
		t.Fatalf("tags: %v", r0.Properties["tags"])
	}
	if len(run.Results) != 3 {
		t.Fatalf("results: %d", len(run.Results))
	}
	res := run.Results[0]
	loc := res.Locations[0].PhysicalLocation
	if res.RuleID != "authz" || res.RuleIndex != 0 || res.Level != "error" ||
		loc.ArtifactLocation.URI != "src/o.ts" || loc.Region.StartLine != 36 ||
		!strings.Contains(res.Message.Text, "CWE-639") ||
		res.PartialFingerprints["primaryLocationLineHash"] == "" {
		t.Fatalf("result: %+v", res)
	}
	if run.Results[1].Level != "warning" || run.Results[1].RuleIndex != 1 {
		t.Fatalf("result 1: %+v", run.Results[1])
	}
}

func TestSARIFDeterministicAndEmpty(t *testing.T) {
	a, _ := FormatSARIF(sample(t), "dev")
	b, _ := FormatSARIF(sample(t), "dev")
	if a != b {
		t.Fatal("not deterministic")
	}
	out, _ := FormatSARIF(nil, "dev")
	if !strings.Contains(out, `"results": []`) || !strings.Contains(out, `"rules": []`) {
		t.Fatalf("empty: %s", out)
	}
}
