package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/triage"
)

const (
	sarifSchema = "https://json.schemastore.org/sarif-2.1.0.json"
	repoURL     = "https://github.com/tiagovilasboas/ai-code-review"
)

// skillFor maps a finding class to the skill doc that explains it.
var skillFor = map[finding.Class]string{
	finding.AuthZ:           "authz-idor",
	finding.Secret:          "secrets-config",
	finding.CORS:            "secrets-config",
	finding.XSS:             "xss-html",
	finding.SSRF:            "ssrf-egress",
	finding.InstallScript:   "supply-chain",
	finding.RegistryHost:    "supply-chain",
	finding.UnpinnedAction:  "supply-chain",
	finding.LockfileRewrite: "supply-chain",
}

// securitySeverity are the GitHub code scanning scores per severity.
var securitySeverity = map[finding.Severity]float64{
	finding.Critical: 9.5, finding.High: 8.0, finding.Medium: 5.5, finding.Low: 3.0,
}

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifRule struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	ShortDescription sarifText      `json:"shortDescription"`
	HelpURI          string         `json:"helpUri,omitempty"`
	Properties       sarifRuleProps `json:"properties"`
}

type sarifRuleProps struct {
	Tags             []string `json:"tags"`
	SecuritySeverity string   `json:"security-severity"`
}

type sarifResult struct {
	RuleID              string             `json:"ruleId"`
	RuleIndex           int                `json:"ruleIndex"`
	Level               string             `json:"level"`
	Message             sarifText          `json:"message"`
	Locations           []sarifLocation    `json:"locations"`
	PartialFingerprints map[string]string  `json:"partialFingerprints"`
	Suppressions        []sarifSuppression `json:"suppressions,omitempty"`
	Properties          *sarifResultProps  `json:"properties,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
		Region struct {
			StartLine int `json:"startLine"`
		} `json:"region"`
	} `json:"physicalLocation"`
}

// FormatSARIF renders a SARIF 2.1.0 log with one rule per finding class present.
func FormatSARIF(findings []finding.Finding, version string) (string, error) {
	return FormatSARIFTriaged(findings, nil, version)
}

// FormatSARIFTriaged renders the findings left after triage, plus every
// dropped finding as a suppressed result and every downgraded one with a
// triage property, each carrying the verdict reason.
func FormatSARIFTriaged(findings []finding.Finding, applied []triage.Applied, version string) (string, error) {
	rules, index := sarifRules(withDropped(findings, applied))
	results := make([]sarifResult, 0, len(findings))
	for _, f := range findings {
		r := sarifResultFor(f, index[f.Class])
		applySarifTriage(&r, f, applied)
		results = append(results, r)
	}
	results = append(results, droppedResults(applied, index)...)
	log := sarifLog{Schema: sarifSchema, Version: "2.1.0", Runs: []sarifRun{{
		Tool: sarifTool{Driver: sarifDriver{
			Name: ToolName, Version: version, InformationURI: repoURL, Rules: rules,
		}},
		Results: results,
	}}}
	out, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

// sarifRules builds the rules sorted by id and a class to rule index map.
func sarifRules(findings []finding.Finding) ([]sarifRule, map[finding.Class]int) {
	tags := map[finding.Class]map[string]bool{}
	worst := map[finding.Class]finding.Severity{}
	for _, f := range findings {
		if tags[f.Class] == nil {
			tags[f.Class] = map[string]bool{"security": true}
			worst[f.Class] = f.Severity
		}
		tags[f.Class][f.CWE] = true
		for _, id := range f.OWASP {
			tags[f.Class][id] = true
		}
		if f.Severity.Rank() < worst[f.Class].Rank() {
			worst[f.Class] = f.Severity
		}
	}
	classes := make([]string, 0, len(tags))
	for c := range tags {
		classes = append(classes, string(c))
	}
	sort.Strings(classes)

	rules := make([]sarifRule, 0, len(classes))
	index := map[finding.Class]int{}
	for i, name := range classes {
		c := finding.Class(name)
		index[c] = i
		r := sarifRule{
			ID: name, Name: c.Label(), ShortDescription: sarifText{c.Label()},
			Properties: sarifRuleProps{
				Tags:             sortedTags(tags[c]),
				SecuritySeverity: fmt.Sprintf("%.1f", securitySeverity[worst[c]]),
			},
		}
		if skill, ok := skillFor[c]; ok {
			r.HelpURI = repoURL + "/blob/main/security/skills/" + skill + ".md"
		}
		rules = append(rules, r)
	}
	return rules, index
}

func sortedTags(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func sarifResultFor(f finding.Finding, ruleIndex int) sarifResult {
	level := "warning"
	if f.Action == finding.Block {
		level = "error"
	}
	var loc sarifLocation
	loc.PhysicalLocation.ArtifactLocation.URI = f.Path
	loc.PhysicalLocation.Region.StartLine = f.Line
	return sarifResult{
		RuleID: string(f.Class), RuleIndex: ruleIndex, Level: level,
		Message:             sarifText{fmt.Sprintf("%s (%s; OWASP: %s)", f.Title, f.CWE, strings.Join(f.OWASP, ", "))},
		Locations:           []sarifLocation{loc},
		PartialFingerprints: map[string]string{"primaryLocationLineHash": fingerprint(f)},
	}
}

// fingerprint is the stable finding ID plus the hash version. The report
// layer has no source text, so the line number stands in for the line content.
func fingerprint(f finding.Finding) string { return finding.ID(f) + ":1" }
