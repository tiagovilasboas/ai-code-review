package tier

import (
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	ok := `{"schemaVersion":1,"high":["a/**"],"medium":[],"low":["*.md"],"alwaysHuman":[],"default":"medium"}`
	if _, err := ParseConfig([]byte(ok)); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	bad := map[string]string{
		"not json":       `{`,
		"wrong version":  `{"schemaVersion":2,"default":"low"}`,
		"missing":        `{"schemaVersion":1}`,
		"bad default":    `{"schemaVersion":1,"default":"urgent"}`,
		"unknown field":  `{"schemaVersion":1,"default":"low","hgih":["x"]}`,
		"empty glob":     `{"schemaVersion":1,"default":"low","high":[""]}`,
		"bad glob":       `{"schemaVersion":1,"default":"low","low":["a**b"]}`,
		"bad char class": `{"schemaVersion":1,"default":"low","medium":["[x"]}`,
	}
	for name, in := range bad {
		if _, err := ParseConfig([]byte(in)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	if _, err := LoadConfig("does/not/exist.json"); err == nil || !strings.Contains(err.Error(), "config") {
		t.Fatalf("err = %v", err)
	}
}

func TestRepoConfigIsValid(t *testing.T) {
	if _, err := LoadConfig("../../.github/review-tiers.json"); err != nil {
		t.Fatalf("repo config: %v", err)
	}
}
