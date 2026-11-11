// Package tier classifies a change into a risk tier from the paths it touches.
//
// The tier is set by the change, not by the author: the highest tier matched by
// any changed path wins. Paths the config marks alwaysHuman, and any CI
// workflow file, always need at least one human approval. The classifier is
// deterministic and never calls a model.
package tier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// SchemaVersion is the only config and output schema this package understands.
const SchemaVersion = 1

// Tier is a risk level. The zero value is invalid.
type Tier string

const (
	Low    Tier = "low"
	Medium Tier = "medium"
	High   Tier = "high"
)

// rank orders tiers; a larger rank is riskier.
func (t Tier) rank() int {
	switch t {
	case Low:
		return 0
	case Medium:
		return 1
	case High:
		return 2
	}
	return -1
}

// Approvals is the number of human approvals a tier needs.
func (t Tier) Approvals() int { return t.rank() }

// maxTier returns the riskier of a and b.
func maxTier(a, b Tier) Tier {
	if b.rank() > a.rank() {
		return b
	}
	return a
}

// Config is the content of .github/review-tiers.json.
type Config struct {
	SchemaVersion  int      `json:"schemaVersion"`
	TrustedAuthors []string `json:"trustedAuthors"`
	High           []string `json:"high"`
	Medium         []string `json:"medium"`
	Low            []string `json:"low"`
	AlwaysHuman    []string `json:"alwaysHuman"`
	Default        Tier     `json:"default"`
}

// workflowGlob is always human, whatever the config says.
const workflowGlob = ".github/workflows/**"

// LoadConfig reads and validates the config file at path.
func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return ParseConfig(b)
}

// ParseConfig decodes and validates config JSON. Unknown fields are rejected so
// a typo in a glob list cannot silently weaken the gate.
func ParseConfig(b []byte) (Config, error) {
	var c Config
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	if err := c.validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c Config) validate() error {
	if c.SchemaVersion != SchemaVersion {
		return fmt.Errorf("config: schemaVersion must be %d, got %d", SchemaVersion, c.SchemaVersion)
	}
	if c.Default.rank() < 0 {
		return fmt.Errorf("config: default must be low, medium or high, got %q", c.Default)
	}
	lists := []struct {
		name  string
		globs []string
	}{{"high", c.High}, {"medium", c.Medium}, {"low", c.Low}, {"alwaysHuman", c.AlwaysHuman}}
	for _, l := range lists {
		for _, g := range l.globs {
			if err := validGlob(g); err != nil {
				return fmt.Errorf("config: %s: %w", l.name, err)
			}
		}
	}
	return nil
}
