package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	scRunKey    = regexp.MustCompile(`^(\s*(?:-\s+)?)run:\s*(.*)$`)
	scBlockHead = regexp.MustCompile(`^[|>][+-]?\d?\s*(?:#.*)?$`)
	scSeparator = regexp.MustCompile(`&&|\|\||[;|]`)
)

type runLine struct {
	line int
	cmd  string
}

func indentOf(s string) int { return len(s) - len(strings.TrimLeft(s, " \t")) }

// addedRunCommands returns the added shell lines of workflow `run:` steps,
// both inline (`run: cmd`) and block scalars (`run: |`). Context lines only
// track where a block starts and ends.
func addedRunCommands(f diff.File) []runLine {
	var out []runLine
	blockIndent := -1
	for _, l := range f.Lines {
		if blockIndent >= 0 {
			if strings.TrimSpace(l.Text) == "" {
				continue
			}
			if indentOf(l.Text) > blockIndent {
				if l.Kind == diff.Added && !IsCommentLine(l.Text) {
					out = append(out, runLine{l.Line, strings.TrimSpace(l.Text)})
				}
				continue
			}
			blockIndent = -1
		}
		m := scRunKey.FindStringSubmatch(l.Text)
		if m == nil {
			continue
		}
		value := strings.TrimSpace(m[2])
		if scBlockHead.MatchString(value) {
			blockIndent = len(m[1])
			continue
		}
		if l.Kind == diff.Added && value != "" {
			out = append(out, runLine{l.Line, strings.Trim(value, `"'`)})
		}
	}
	return out
}

// installRewritesLockfile reports whether any command of a shell line is a
// package install that may update the lockfile (not npm ci, a frozen/immutable
// install, a global tool install, or an install with lockfile writes off).
func installRewritesLockfile(cmd string) bool {
	if i := strings.Index(cmd, " #"); i >= 0 {
		cmd = cmd[:i]
	}
	for _, seg := range scSeparator.Split(cmd, -1) {
		if segmentRewritesLockfile(strings.Fields(seg)) {
			return true
		}
	}
	return false
}

func segmentRewritesLockfile(tok []string) bool {
	for len(tok) > 0 && (tok[0] == "sudo" || strings.Contains(tok[0], "=")) {
		tok = tok[1:]
	}
	if len(tok) == 0 {
		return false
	}
	sub, flags := "", map[string]bool{}
	for _, t := range tok[1:] {
		if strings.HasPrefix(t, "-") {
			flags[t] = true
		} else if sub == "" {
			sub = t
		}
	}
	switch tok[0] {
	case "npm":
		return (sub == "install" || sub == "i") && !hasAny(flags, "-g", "--global", "--location=global",
			"--no-package-lock", "--package-lock=false", "--no-save", "--dry-run", "--frozen-lockfile")
	case "yarn":
		bare := sub == "" || sub == "install"
		return bare && !hasAny(flags, "--frozen-lockfile", "--immutable", "--pure-lockfile")
	case "bun":
		return (sub == "install" || sub == "i") && !hasAny(flags, "--frozen-lockfile", "-g", "--global")
	case "pnpm":
		return (sub == "install" || sub == "i") && hasAny(flags, "--no-frozen-lockfile", "--fix-lockfile")
	}
	return false
}

func hasAny(set map[string]bool, keys ...string) bool {
	for _, k := range keys {
		if set[k] {
			return true
		}
	}
	return false
}
