package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/rules"
)

func TestFormatJSONAndSARIF(t *testing.T) {
	d := writeDiff(t)
	rs := []rules.Rule{blockRule}
	for _, format := range []string{"json", "sarif"} {
		for _, args := range [][]string{{"--format", format, d}, {"--format=" + format, "--", d}} {
			code, out, _ := exec(args, rs)
			if code != 1 || !json.Valid([]byte(out)) {
				t.Fatalf("%v: %d %q", args, code, out)
			}
		}
	}
}

func TestOutputFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "r.sarif")
	code, stdout, _ := exec([]string{"--format", "sarif", "--output", out, writeDiff(t)}, []rules.Rule{blockRule})
	b, err := os.ReadFile(out)
	if code != 1 || stdout != "" || err != nil || !strings.Contains(string(b), "2.1.0") {
		t.Fatalf("%d %q %v", code, stdout, err)
	}
}

func TestOutputUnwritable(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "no", "r.json")
	code, _, errs := exec([]string{"--output", bad, writeDiff(t)}, nil)
	if code != 2 || !strings.HasPrefix(errs, "error: cannot write "+bad) {
		t.Fatalf("%d %q", code, errs)
	}
}

func TestUsageErrors(t *testing.T) {
	d := writeDiff(t)
	cases := map[string][]string{
		"unknown format":      {"--format", "xml", d},
		"from without to":     {"--git", "--from", "a"},
		"to without from":     {"--git", "--to", "b"},
		"commit with from":    {"--git", "--commit", "x", "--from", "a", "--to", "b"},
		"git with file":       {"--git", d},
		"range without git":   {"--from", "a", "--to", "b", d},
		"unknown flag":        {"--nope", d},
		"missing value":       {"--format"},
		"two files":           {d, d},
		"value on bool flag":  {"--git=1"},
		"format without file": {"--format", "json"},
	}
	for name, args := range cases {
		code, out, errs := exec(args, nil)
		if code != 2 || out != "" || !strings.Contains(errs, "Usage:") {
			t.Errorf("%s: %d %q %q", name, code, out, errs)
		}
	}
}

func TestUsageListsAllFlags(t *testing.T) {
	for _, f := range []string{"--format", "--output", "--git", "--from", "--to", "--commit", "--dir", "--version", "--"} {
		if !strings.Contains(usageText, f) {
			t.Errorf("usage misses %s", f)
		}
	}
}

func TestGitSourceErrorsExitTwo(t *testing.T) {
	dir := t.TempDir() // not a git repository
	for _, args := range [][]string{
		{"--git", "--dir", dir},
		{"--git", "--from", "a", "--to", "b", "--dir", dir},
		{"--git", "--commit", "abc", "--dir", dir},
	} {
		code, out, errs := exec(args, nil)
		if code != 2 || out != "" || !strings.HasPrefix(errs, "error: ") || strings.Count(errs, "\n") != 1 {
			t.Errorf("%v: %d %q %q", args, code, out, errs)
		}
	}
}

func TestVersion(t *testing.T) {
	code, out, _ := exec([]string{"--version"}, nil)
	if code != 0 || out != "review dev\n" {
		t.Fatalf("%d %q", code, out)
	}
}
