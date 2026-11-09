// Package bench measures the deterministic review engine against a labelled
// corpus of diffs: precision, recall and F1 per finding class.
package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// Key identifies one finding for matching: new-file path, line and class.
type Key struct {
	Path  string `json:"path"`
	Line  int    `json:"line"`
	Class string `json:"class"`
}

// Case is one labelled diff. Empty Expected means the engine must stay silent.
type Case struct {
	ID          string
	Description string
	Source      string
	Diff        string
	Expected    []Key
}

type expectedFile struct {
	Description string `json:"description"`
	Source      string `json:"source"`
	Findings    []Key  `json:"findings"`
}

var knownClasses = map[string]bool{
	string(finding.AuthZ): true, string(finding.Secret): true, string(finding.CORS): true,
	string(finding.XSS): true, string(finding.SSRF): true, string(finding.InstallScript): true,
	string(finding.RegistryHost): true, string(finding.UnpinnedAction): true,
	string(finding.LockfileRewrite): true, string(finding.UntrustedDiff): true,
	string(finding.SQLInjection): true, string(finding.CmdInjection): true,
	string(finding.PathTraversal): true,
}

// LoadCorpus reads every <dir>/<case-id>/{change.diff,expected.json}, sorted by id.
// Any malformed case is an error: a silently skipped case would skew the numbers.
func LoadCorpus(dir string) ([]Case, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read corpus: %w", err)
	}
	var cases []Case
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := loadCase(filepath.Join(dir, e.Name()), e.Name())
		if err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("corpus %s has no cases", dir)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	return cases, nil
}

func loadCase(dir, id string) (Case, error) {
	diffBytes, err := os.ReadFile(filepath.Join(dir, "change.diff"))
	if err != nil {
		return Case{}, fmt.Errorf("case %s: %w", id, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "expected.json"))
	if err != nil {
		return Case{}, fmt.Errorf("case %s: %w", id, err)
	}
	var exp expectedFile
	if err := json.Unmarshal(raw, &exp); err != nil {
		return Case{}, fmt.Errorf("case %s: expected.json: %w", id, err)
	}
	for _, k := range exp.Findings {
		if k.Path == "" || k.Line < 1 || !knownClasses[k.Class] {
			return Case{}, fmt.Errorf("case %s: invalid expected finding %+v", id, k)
		}
	}
	return Case{ID: id, Description: exp.Description, Source: exp.Source, Diff: string(diffBytes), Expected: exp.Findings}, nil
}
