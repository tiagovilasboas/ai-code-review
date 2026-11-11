package tier

import (
	"flag"
	"fmt"
	"io"

	"github.com/tiagovilasboas/ai-code-review/internal/source"
)

const usage = `Usage:
  tier [--config FILE] [--format text|json] [--author LOGIN] <diff-file | ->
  tier [--config FILE] [--format text|json] [--author LOGIN] --from REF --to REF [--dir DIR]
  tier [--config FILE] [--format text|json] [--author LOGIN] --incomplete REASON

--incomplete means the change set could not be read in full: the result is high
and REASON is recorded (fail closed). No diff input is read.

Exit codes: 0 classified, 2 usage, read, git or config error.
`

// Main runs the tier CLI and returns the process exit code.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("tier", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	cfgPath := fs.String("config", ".github/review-tiers.json", "tier config file")
	format := fs.String("format", "text", "text or json")
	author := fs.String("author", "", "PR author login")
	from := fs.String("from", "", "base ref")
	to := fs.String("to", "", "head ref")
	incomplete := fs.String("incomplete", "", "change set unreadable: classify high with this reason")
	dir := fs.String("dir", ".", "repository directory for --from/--to")
	if err := fs.Parse(args); err != nil {
		return fail(stderr, err.Error())
	}
	if *format != "text" && *format != "json" {
		return fail(stderr, fmt.Sprintf("unknown format %q", *format))
	}
	var diff string
	if *incomplete == "" {
		var err error
		if diff, err = readDiff(fs.Args(), *from, *to, *dir, stdin); err != nil {
			return fail(stderr, err.Error())
		}
	} else if *from != "" || *to != "" || len(fs.Args()) > 0 {
		return fail(stderr, "--incomplete replaces every diff input")
	}
	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		return fail(stderr, err.Error())
	}
	res := Classify(cfg, ChangedPaths(diff), *author)
	if *incomplete != "" {
		res = MarkIncomplete(res, *incomplete)
	}
	if *format == "json" {
		out, err := JSON(res)
		if err != nil {
			return fail(stderr, err.Error())
		}
		fmt.Fprint(stdout, out)
		return 0
	}
	fmt.Fprint(stdout, Text(res))
	return 0
}

func readDiff(pos []string, from, to, dir string, stdin io.Reader) (string, error) {
	switch {
	case from != "" || to != "":
		if from == "" || to == "" || len(pos) > 0 {
			return "", fmt.Errorf("--from and --to go together and replace the diff argument")
		}
		return source.Range(dir, from, to)
	case len(pos) == 1 && pos[0] == "-":
		b, err := io.ReadAll(stdin)
		return string(b), err
	case len(pos) == 1:
		return source.File(pos[0])
	}
	return "", fmt.Errorf("give one diff file, '-' for stdin, or --from and --to")
}

func fail(w io.Writer, msg string) int {
	fmt.Fprintf(w, "tier: %s\n\n%s", msg, usage)
	return 2
}
