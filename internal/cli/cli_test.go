package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/rules"
)

func exec(args []string, rs []rules.Rule) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, rs, &out, &errb)
	return code, out.String(), errb.String()
}

func writeDiff(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "a.diff")
	text := "diff --git a/f b/f\n+++ b/f\n@@ -1 +1 @@\n+x\n"
	if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func blockRule(f diff.File) []finding.Finding {
	x, _ := finding.New(finding.Input{
		Title: "T", Class: finding.XSS, Path: f.Path, Line: 1, CWE: "CWE-79",
		OWASP: finding.LabelsFor(finding.XSS), Severity: finding.High,
	})
	return []finding.Finding{x}
}

func TestUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"--"}} {
		code, out, errs := exec(args, nil)
		if code != 2 || out != "" || errs != "Usage: review <diff-file>\n" {
			t.Fatalf("%v: %d %q %q", args, code, out, errs)
		}
	}
}

func TestUnreadable(t *testing.T) {
	code, _, errs := exec([]string{"/nonexistent/x.diff"}, nil)
	want := "error: cannot read /nonexistent/x.diff: no such file or directory\n"
	if code != 2 || errs != want {
		t.Fatalf("%d %q", code, errs)
	}
}

func TestCleanDiff(t *testing.T) {
	code, out, _ := exec([]string{"--", writeDiff(t)}, nil)
	if code != 0 || out != "no evidence-based findings\n" {
		t.Fatalf("%d %q", code, out)
	}
}

func TestBlockExitsOne(t *testing.T) {
	code, out, _ := exec([]string{writeDiff(t)}, []rules.Rule{blockRule})
	if code != 1 || !strings.HasPrefix(out, "Decision: DO NOT MERGE\n") {
		t.Fatalf("%d %q", code, out)
	}
}
