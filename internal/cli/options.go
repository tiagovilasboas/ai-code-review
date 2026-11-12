package cli

import (
	"fmt"
	"strings"
)

const usageText = `Usage:
  review [--format text|json|sarif] [--output FILE] [--triage FILE] <diff-file>
  review [--format text|json|sarif] [--output FILE] [--triage FILE] --git [--from REF --to REF | --commit SHA] [--dir DIR]
  review --version

Flags:
  --format FMT    report format: text (default), json or sarif
  --output FILE   write the report to FILE instead of stdout
  --triage FILE   apply external triage verdicts (JSON) that may only drop or
                  downgrade findings; an invalid file is exit 2, no report
  --git           read the diff from git instead of a file
                  (no range: workspace changes vs HEAD)
  --from REF      with --to: changes of --to since its merge-base with REF
  --to REF        see --from
  --commit SHA    review a single commit
  --dir DIR       repository directory for --git (default ".")
  --version       print the version
  --              end of flags; the next argument is the diff file

Exit codes: 0 no BLOCK finding, 1 any BLOCK finding, 2 usage, read or git error.
`

// options is the parsed command line.
type options struct {
	format, output     string
	triage             string
	git, version, help bool
	from, to, commit   string
	dir, file          string
}

// parseArgs parses flags in --flag value and --flag=value form. The returned
// error is a usage problem; an empty message means print the usage only.
func parseArgs(args []string) (options, error) {
	o := options{format: "text", dir: "."}
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(a, "--") {
			positional = append(positional, a)
			continue
		}
		name, val, hasVal := strings.Cut(a[2:], "=")
		switch name {
		case "git", "version", "help":
			if hasVal {
				return o, usageErr("flag --%s takes no value", name)
			}
			o.git = o.git || name == "git"
			o.version = o.version || name == "version"
			o.help = o.help || name == "help"
			continue
		}
		dst := o.valueFlag(name)
		if dst == nil {
			return o, usageErr("unknown flag %s", a)
		}
		if !hasVal {
			if i+1 >= len(args) {
				return o, usageErr("flag --%s needs a value", name)
			}
			i++
			val = args[i]
		}
		*dst = val
	}
	return o, o.finish(positional)
}

func (o *options) valueFlag(name string) *string {
	switch name {
	case "format":
		return &o.format
	case "output":
		return &o.output
	case "triage":
		return &o.triage
	case "from":
		return &o.from
	case "to":
		return &o.to
	case "commit":
		return &o.commit
	case "dir":
		return &o.dir
	}
	return nil
}

// finish validates the flag combination against the CLI contract.
func (o *options) finish(positional []string) error {
	if o.version || o.help {
		return nil
	}
	switch o.format {
	case "text", "json", "sarif":
	default:
		return usageErr("unknown format %q (want text, json or sarif)", o.format)
	}
	if len(positional) > 1 {
		return usageErr("expected one diff file, got %d", len(positional))
	}
	if len(positional) == 1 {
		o.file = positional[0]
	}
	if !o.git {
		return o.finishFile()
	}
	switch {
	case o.file != "":
		return usageErr("--git cannot be combined with a diff file")
	case (o.from == "") != (o.to == ""):
		return usageErr("--from and --to must be used together")
	case o.commit != "" && o.from != "":
		return usageErr("--commit cannot be combined with --from/--to")
	}
	return nil
}

func (o *options) finishFile() error {
	switch {
	case o.from != "" || o.to != "" || o.commit != "":
		return usageErr("--from, --to and --commit need --git")
	case o.file == "":
		return usageErr("")
	}
	return nil
}

type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

func usageErr(format string, a ...any) error {
	return usageError{fmt.Sprintf(format, a...)}
}
