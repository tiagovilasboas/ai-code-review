package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoldenParity runs every examples/*.diff through the Go CLI and compares
// the report byte for byte with testdata/golden. Goldens of the JS-era
// fixtures were captured from the JS engine; the injection fixtures
// (sql-injection, command-injection, path-traversal) are Go-only and their
// goldens come from the Go CLI.
func TestGoldenParity(t *testing.T) {
	root := filepath.Join("..", "..")
	diffs, err := filepath.Glob(filepath.Join(root, "examples", "*.diff"))
	if err != nil || len(diffs) == 0 {
		t.Fatalf("no example diffs found: %v", err)
	}
	for _, diffPath := range diffs {
		name := strings.TrimSuffix(filepath.Base(diffPath), ".diff")
		t.Run(name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join(root, "testdata", "golden", name+".txt"))
			if err != nil {
				t.Fatalf("missing golden for %s: %v", name, err)
			}
			var stdout, stderr bytes.Buffer
			code := Main([]string{diffPath}, &stdout, &stderr)
			if got := stdout.String(); got != string(want) {
				t.Errorf("report mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
			}
			wantCode := 0
			if strings.Contains(string(want), "Action: BLOCK") {
				wantCode = 1
			}
			if code != wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %s)", code, wantCode, stderr.String())
			}
		})
	}
}
