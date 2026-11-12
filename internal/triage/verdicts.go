// Package triage applies an optional, externally produced set of verdicts to
// the engine's findings. The engine never calls a model: a verdicts file is
// plain data, and Apply enforces in code that it can only drop or lower
// findings, never add or raise one.
package triage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// SchemaVersion is the only verdicts schema this build understands.
const SchemaVersion = 1

// maxReason bounds a reason so a verdict cannot smuggle a payload into reports.
const maxReason = 500

// Verb is what a verdict does to a finding.
type Verb string

const (
	Drop      Verb = "drop"
	Downgrade Verb = "downgrade"
)

// Verdict is one decision about one existing finding, named by its ID.
type Verdict struct {
	ID       string           `json:"id"`
	Action   Verb             `json:"action"`
	Severity finding.Severity `json:"severity,omitempty"`
	Reason   string           `json:"reason"`
}

// File is the verdicts document.
type File struct {
	SchemaVersion int       `json:"schemaVersion"`
	Verdicts      []Verdict `json:"verdicts"`
}

// Parse decodes and validates a verdicts file strictly: unknown fields (a
// smuggled "findings" list, for example), trailing data, a wrong schema
// version, duplicate ids and malformed verdicts are all errors.
func Parse(data []byte) (File, error) {
	var f File
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return File{}, fmt.Errorf("malformed verdicts file: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return File{}, errors.New("malformed verdicts file: trailing data after the JSON document")
	}
	if f.SchemaVersion != SchemaVersion {
		return File{}, fmt.Errorf("unsupported verdicts schemaVersion %d (want %d)", f.SchemaVersion, SchemaVersion)
	}
	seen := map[string]bool{}
	for i, v := range f.Verdicts {
		if err := v.validate(); err != nil {
			return File{}, fmt.Errorf("verdict %d: %w", i, err)
		}
		if seen[v.ID] {
			return File{}, fmt.Errorf("verdict %d: duplicate id %s", i, v.ID)
		}
		seen[v.ID] = true
	}
	return f, nil
}

func (v Verdict) validate() error {
	if v.ID == "" {
		return errors.New("missing id")
	}
	if v.Reason == "" {
		return fmt.Errorf("%s: a reason is required", v.ID)
	}
	if len(v.Reason) > maxReason {
		return fmt.Errorf("%s: reason longer than %d bytes", v.ID, maxReason)
	}
	switch v.Action {
	case Drop:
		if v.Severity != "" {
			return fmt.Errorf("%s: drop takes no severity", v.ID)
		}
	case Downgrade:
		switch v.Severity {
		case finding.Critical, finding.High, finding.Medium, finding.Low:
		default:
			return fmt.Errorf("%s: downgrade needs a severity of CRITICAL, HIGH, MEDIUM or LOW", v.ID)
		}
	default:
		return fmt.Errorf("%s: unknown action %q (want drop or downgrade)", v.ID, v.Action)
	}
	return nil
}
