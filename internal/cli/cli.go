// Package cli holds the argument handling of the review command.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tiagovilasboas/ai-code-review/internal/report"
	"github.com/tiagovilasboas/ai-code-review/internal/review"
	"github.com/tiagovilasboas/ai-code-review/internal/rules"
)

// Main runs the CLI with the default rule set and returns the exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	return run(args, review.Default, stdout, stderr)
}

func run(args []string, rs []rules.Rule, stdout, stderr io.Writer) int {
	path := ""
	for _, a := range args {
		if a != "--" {
			path = a
			break
		}
	}
	if path == "" {
		fmt.Fprint(stderr, "Usage: review <diff-file>\n")
		return 2
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot read %s: %s\n", path, reason(err))
		return 2
	}

	findings := review.Run(string(data), rs)
	fmt.Fprint(stdout, report.FormatText(findings))
	return report.ExitCode(findings)
}

// reason strips the "open <path>:" prefix of a PathError.
func reason(err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}
