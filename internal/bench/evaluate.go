package bench

import (
	"sort"

	"github.com/tiagovilasboas/ai-code-review/internal/review"
)

// Analyzer turns a diff into the keys the engine reports.
type Analyzer func(diffText string) []Key

// DefaultAnalyzer runs the production rule set.
func DefaultAnalyzer(diffText string) []Key {
	var keys []Key
	for _, f := range review.Run(diffText, review.Default) {
		keys = append(keys, Key{Path: f.Path, Line: f.Line, Class: string(f.Class)})
	}
	return keys
}

// CaseResult is the outcome for one corpus case.
type CaseResult struct {
	ID     string `json:"id"`
	TP     int    `json:"tp"`
	Missed []Key  `json:"missed,omitempty"`
	Extra  []Key  `json:"extra,omitempty"`
}

// ClassResult is the confusion count for one finding class.
type ClassResult struct {
	Class string `json:"class"`
	Counts
}

// Report is the full benchmark outcome.
type Report struct {
	Cases       int           `json:"cases"`
	CleanCases  int           `json:"cleanCases"`
	CleanSilent int           `json:"cleanSilent"`
	Overall     Counts        `json:"overall"`
	Classes     []ClassResult `json:"classes"`
	CaseResults []CaseResult  `json:"caseResults"`
}

// Evaluate runs the analyzer on every case and tallies the confusion counts.
// A false positive counts against the class the engine reported; a false
// negative against the class that was expected.
func Evaluate(cases []Case, analyze Analyzer) Report {
	r := Report{Cases: len(cases)}
	byClass := map[string]*Counts{}
	for name := range knownClasses {
		byClass[name] = &Counts{}
	}
	bump := func(class string) *Counts {
		if byClass[class] == nil {
			byClass[class] = &Counts{}
		}
		return byClass[class]
	}
	for _, c := range cases {
		hit, missed, extra := Match(c.Expected, analyze(c.Diff))
		r.CaseResults = append(r.CaseResults, CaseResult{ID: c.ID, TP: len(hit), Missed: missed, Extra: extra})
		for _, k := range hit {
			bump(k.Class).TP++
		}
		for _, k := range missed {
			bump(k.Class).FN++
		}
		for _, k := range extra {
			bump(k.Class).FP++
		}
		r.Overall.TP += len(hit)
		r.Overall.FN += len(missed)
		r.Overall.FP += len(extra)
		if len(c.Expected) == 0 {
			r.CleanCases++
			if len(extra) == 0 {
				r.CleanSilent++
			}
		}
	}
	names := make([]string, 0, len(byClass))
	for name := range byClass {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		r.Classes = append(r.Classes, ClassResult{Class: name, Counts: *byClass[name]})
	}
	return r
}
