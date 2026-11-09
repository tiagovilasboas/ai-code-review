// Package cli holds the argument handling of the review command.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
	"github.com/tiagovilasboas/ai-code-review/internal/report"
	"github.com/tiagovilasboas/ai-code-review/internal/review"
	"github.com/tiagovilasboas/ai-code-review/internal/rules"
	"github.com/tiagovilasboas/ai-code-review/internal/source"
	"github.com/tiagovilasboas/ai-code-review/internal/version"
)

// Main runs the CLI with the default rule set and returns the exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	return run(args, review.Default, stdout, stderr)
}

func run(args []string, rs []rules.Rule, stdout, stderr io.Writer) int {
	o, err := parseArgs(args)
	if err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintf(stderr, "error: %s\n", msg)
		}
		fmt.Fprint(stderr, usageText)
		return 2
	}
	switch {
	case o.help:
		fmt.Fprint(stdout, usageText)
		return 0
	case o.version:
		fmt.Fprintf(stdout, "review %s\n", version.Version)
		return 0
	}

	text, err := readDiff(o)
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", err)
		return 2
	}

	findings := review.Run(text, rs)
	out, err := render(o.format, findings)
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", err)
		return 2
	}
	if o.output == "" {
		fmt.Fprint(stdout, out)
	} else if err := os.WriteFile(o.output, []byte(out), 0o644); err != nil {
		fmt.Fprintf(stderr, "error: cannot write %s: %s\n", o.output, reason(err))
		return 2
	}
	return report.ExitCode(findings)
}

// readDiff loads the unified diff from the file or the git source.
func readDiff(o options) (string, error) {
	switch {
	case !o.git:
		data, err := source.File(o.file)
		if err != nil {
			return "", fmt.Errorf("cannot read %s: %s", o.file, reason(err))
		}
		return data, nil
	case o.commit != "":
		return source.Commit(o.dir, o.commit)
	case o.from != "":
		return source.Range(o.dir, o.from, o.to)
	}
	return source.Workspace(o.dir)
}

func render(format string, findings []finding.Finding) (string, error) {
	switch format {
	case "json":
		return report.FormatJSON(findings, version.Version)
	case "sarif":
		return report.FormatSARIF(findings, version.Version)
	}
	return report.FormatText(findings), nil
}

// reason strips the "open <path>:" prefix of a PathError.
func reason(err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}
