package tier

import "testing"

func testConfig() Config {
	return Config{
		SchemaVersion:  1,
		TrustedAuthors: []string{"owner"},
		High:           []string{"internal/rules/**", "action.yml"},
		Medium:         []string{"internal/**", "skills/**"},
		Low:            []string{"docs/**", "*.md"},
		AlwaysHuman:    []string{"action.yml", "CODEOWNERS"},
		Default:        Medium,
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name        string
		paths       []string
		author      string
		tier        Tier
		humans      int
		alwaysHuman bool
	}{
		{"empty diff is low", nil, "owner", Low, 0, false},
		{"docs only", []string{"docs/a.md", "README.md"}, "owner", Low, 0, false},
		{"medium", []string{"internal/x/y.go"}, "owner", Medium, 1, false},
		{"highest wins", []string{"docs/a.md", "internal/rules/r.go", "internal/z.go"}, "owner", High, 2, false},
		{"unknown path uses default", []string{"mystery/file.bin"}, "owner", Medium, 1, false},
		{"markdown inside medium dir is medium", []string{"skills/a/SKILL.md"}, "owner", Medium, 1, false},
		{"workflow is always human", []string{".github/workflows/ci.yml"}, "owner", Medium, 1, true},
		{"always human forces a person on a low tier", []string{"CODEOWNERS"}, "owner", Medium, 1, true},
		{"always human on high keeps 2", []string{"action.yml"}, "owner", High, 2, true},
		{"deleted workflow still counts", []string{".github/workflows/old.yml"}, "owner", Medium, 1, true},
		{"untrusted author on docs is medium", []string{"docs/a.md"}, "stranger", Medium, 1, false},
		{"untrusted author on high stays high", []string{"internal/rules/r.go"}, "stranger", High, 2, false},
		{"trusted check is case-insensitive", []string{"docs/a.md"}, "OWNER", Low, 0, false},
		{"no author skips the rule", []string{"docs/a.md"}, "", Low, 0, false},
		{"untrusted author on empty diff", nil, "stranger", Medium, 1, false},
	}
	for _, c := range cases {
		r := Classify(testConfig(), c.paths, c.author)
		if r.Tier != c.tier || r.HumanApprovals != c.humans || r.AlwaysHuman != c.alwaysHuman {
			t.Errorf("%s: got %s/%d/%v, want %s/%d/%v", c.name, r.Tier, r.HumanApprovals, r.AlwaysHuman, c.tier, c.humans, c.alwaysHuman)
		}
		if r.Reasons == nil {
			t.Errorf("%s: Reasons must be a non-nil slice", c.name)
		}
	}
}

func TestClassifyEmptyTrustedListSkipsAuthorRule(t *testing.T) {
	cfg := testConfig()
	cfg.TrustedAuthors = nil
	if r := Classify(cfg, []string{"docs/a.md"}, "anyone"); r.Tier != Low {
		t.Fatalf("tier = %s, want low", r.Tier)
	}
}

func TestClassifyReasonsOrderedByTier(t *testing.T) {
	r := Classify(testConfig(), []string{"docs/a.md", "internal/rules/r.go"}, "owner")
	if r.Reasons[0].Tier != High || r.Reasons[0].Path != "internal/rules/r.go" {
		t.Fatalf("first reason = %+v", r.Reasons[0])
	}
}
