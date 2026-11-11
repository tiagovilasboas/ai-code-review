package tier

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cfgJSON = `{"schemaVersion":1,"trustedAuthors":["owner"],"high":["internal/rules/**"],
"medium":["internal/**"],"low":["docs/**"],"alwaysHuman":[],"default":"medium"}`

func run(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := Main(args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func writeCfg(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tiers.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMainStdinJSON(t *testing.T) {
	cfg := writeCfg(t, cfgJSON)
	diff := "diff --git a/internal/rules/r.go b/internal/rules/r.go\n--- a/internal/rules/r.go\n+++ b/internal/rules/r.go\n@@ -1 +1 @@\n-a\n+b\n"
	code, out, _ := run(t, diff, "--config", cfg, "--format", "json", "--author", "owner", "-")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var r Result
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatal(err)
	}
	if r.SchemaVersion != 1 || r.Tier != High || r.HumanApprovals != 2 {
		t.Fatalf("result = %+v", r)
	}
}

func TestMainEmptyDiffIsLow(t *testing.T) {
	cfg := writeCfg(t, cfgJSON)
	code, out, _ := run(t, "", "--config", cfg, "-")
	if code != 0 || !strings.Contains(out, "Risk tier: low") {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestMainErrorsExitTwo(t *testing.T) {
	good := writeCfg(t, cfgJSON)
	bad := writeCfg(t, `{"schemaVersion":9}`)
	cases := map[string][]string{
		"bad config":     {"--config", bad, "-"},
		"missing config": {"--config", "nope.json", "-"},
		"no input":       {"--config", good},
		"bad format":     {"--config", good, "--format", "xml", "-"},
		"only from":      {"--config", good, "--from", "main"},
		"unknown flag":   {"--bogus"},
		"missing file":   {"--config", good, "nope.diff"},
		"option ref":     {"--config", good, "--from", "--evil", "--to", "HEAD", "--dir", t.TempDir()},
	}
	for name, args := range cases {
		if code, _, errs := run(t, "", args...); code != 2 || !strings.Contains(errs, "tier:") {
			t.Errorf("%s: code %d stderr %q", name, code, errs)
		}
	}
}

func TestMainRangeNeedsNoPositional(t *testing.T) {
	cfg := writeCfg(t, cfgJSON)
	if code, _, _ := run(t, "", "--config", cfg, "--from", "a", "--to", "b", "x.diff"); code != 2 {
		t.Fatalf("range plus file accepted, code %d", code)
	}
}
