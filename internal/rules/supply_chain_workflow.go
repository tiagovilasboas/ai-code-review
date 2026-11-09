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
	// docker://image@sha256:<64 hex> is immutable; a tag alone is not.
	scDockerDigest = regexp.MustCompile(`(?i)^docker://\S+@sha256:[0-9a-f]{64}$`)
)

func workflowFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) {
			continue
		}
		if m := scUses.FindStringSubmatch(l.Text); m != nil {
			ref := strings.Trim(m[1], `"'`)
			if !strings.HasPrefix(ref, "./") && !scFirstParty.MatchString(ref) && !scPinnedSHA.MatchString(ref) &&
				!scDockerDigest.MatchString(ref) {
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
