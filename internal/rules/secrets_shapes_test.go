package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// secretShape is one added line and the number of secret findings it must raise.
type secretShape struct {
	name, path, line string
	want             int
}

func runSecretShapes(t *testing.T, cases []secretShape) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(Secrets(fileOf(c.path, 1, c.line)), finding.Secret)
			if got != c.want {
				t.Fatalf("got %d secret findings, want %d", got, c.want)
			}
		})
	}
}
