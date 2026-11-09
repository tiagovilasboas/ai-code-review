package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	scFirstParty   = regexp.MustCompile(`^(?:actions|github)/`)
	scPinnedSHA    = regexp.MustCompile(`(?i)@[0-9a-f]{40}$`)
	scWorkflowFile = regexp.MustCompile(`\.ya?ml$`)
	scUses         = regexp.MustCompile(`\buses:\s*(\S+)`)
)

func workflowFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) {
			continue
		}
		if m := scUses.FindStringSubmatch(l.Text); m != nil {
			if !strings.HasPrefix(m[1], "./") && !scFirstParty.MatchString(m[1]) && !scPinnedSHA.MatchString(m[1]) {
				out = add(out, scInput(f, l.Line, finding.UnpinnedAction, "Unpinned third-party GitHub Action", "CWE-829", finding.High))
			}
		}
	}
	for _, r := range addedRunCommands(f) {
		if installRewritesLockfile(r.cmd) {
			out = add(out, scInput(f, r.line, finding.LockfileRewrite, "CI install can rewrite the lockfile", "CWE-1104", finding.Medium))
		}
	}
	return out
}
