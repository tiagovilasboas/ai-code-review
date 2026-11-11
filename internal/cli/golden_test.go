package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoldenReports runs every security/examples/*.diff through the CLI and
// compares the report byte for byte with testdata/golden/<name>.txt.
func TestGoldenReports(t *testing.T) {
	root := filepath.Join("..", "..")
	diffs, err := filepath.Glob(filepath.Join(root, "security", "examples", "*.diff"))
	if err != nil || len(diffs) == 0 {
		t.Fatalf("no example diffs found: %v", err)
	}
	for _, diffPath := range diffs {
		name := strings.TrimSuffix(filepath.Base(diffPath), ".diff")
		t.Run(name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", "golden", name+".txt"))
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
