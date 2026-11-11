package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	tfResource    = regexp.MustCompile(`^\s*resource\s+"([\w]+)"`)
	tfRulesetType = "cloudflare_ruleset"
	tfManaged     = regexp.MustCompile(`http_request_firewall_managed|kind\s*=\s*"managed"`)
	tfRateLimit   = regexp.MustCompile(`phase\s*=\s*"http_ratelimit"`)
	tfDisabled    = regexp.MustCompile(`^\s*enabled\s*=\s*false\b`)
	tfSkip        = regexp.MustCompile(`^\s*action\s*=\s*"skip"`)
	tfOverrides   = regexp.MustCompile(`\boverrides\b`)
	tfSkipAll     = regexp.MustCompile(`^\s*(?:ruleset|phases)\s*=`)
)

// rulesetBlock is the run of diff lines belonging to one resource block.
type rulesetBlock struct {
	ruleset bool
	lines   []diff.Line
}

// rulesetBlocks splits the visible diff lines at each `resource` header. Lines
// before the first header belong to no known block and are never judged.
func rulesetBlocks(lines []diff.Line) []rulesetBlock {
	var out []rulesetBlock
	cur := -1
	for _, l := range lines {
		if m := tfResource.FindStringSubmatch(l.Text); m != nil {
			out = append(out, rulesetBlock{ruleset: m[1] == tfRulesetType})
			cur = len(out) - 1
		}
		if cur >= 0 {
			out[cur].lines = append(out[cur].lines, l)
		}
	}
	return out
}

// terraformRulesetFindings flags, in a cloudflare_ruleset whose phase is visible
// in the diff: a managed WAF ruleset disabled or skipped as a whole, a rate
// limit rule disabled, and a rate limit ruleset removed (deleted phase line,
// anchored on the deletion). A skip that lists specific rules is a tuning
// exception and stays silent.
func terraformRulesetFindings(f diff.File) []finding.Finding {
	if !tfFile.MatchString(f.Path) {
		return nil
	}
	var out []finding.Finding
	for _, b := range rulesetBlocks(f.Lines) {
		if !b.ruleset {
			continue
		}
		text := joinLines(b.lines)
		managed, limited := tfManaged.MatchString(text), tfRateLimit.MatchString(text)
		overridden := false // enabled = false inside overrides is per-rule tuning
		for i, l := range b.lines {
			overridden = overridden || tfOverrides.MatchString(l.Text)
			if l.Kind != diff.Added || IsCommentLine(l.Text) {
				continue
			}
			switch {
			case managed && (tfDisabled.MatchString(l.Text) && !overridden || tfSkip.MatchString(l.Text) && skipsWholeRuleset(b.lines[i:])):
				out = add(out, cfConfig(f, l.Line, "Managed WAF ruleset disabled", "CWE-693", finding.Medium))
			case limited && tfDisabled.MatchString(l.Text):
				out = add(out, cfConfig(f, l.Line, "Rate limiting rule disabled", "CWE-770", finding.Medium, "API4:2023"))
			}
		}
	}
	return append(out, removedRateLimit(f)...)
}

// skipsWholeRuleset reports a skip action followed (within the block) by a
// ruleset or phases target rather than a list of rules.
func skipsWholeRuleset(rest []diff.Line) bool {
	for i, l := range rest {
		if i > 0 && i < 10 && tfSkipAll.MatchString(l.Text) {
			return true
		}
	}
	return false
}

// removedRateLimit flags a deleted rate limit phase line. A phase line that is
// added elsewhere in the file means the ruleset moved or was rewritten.
func removedRateLimit(f diff.File) []finding.Finding {
	for _, l := range f.Added {
		if tfRateLimit.MatchString(l.Text) {
			return nil
		}
	}
	var out []finding.Finding
	for _, r := range f.Removed {
		if tfRateLimit.MatchString(r.Text) && !IsCommentLine(r.Text) {
			out = add(out, cfConfig(f, r.Line, "Rate limiting ruleset removed", "CWE-770", finding.Medium, "API4:2023"))
		}
	}
	return out
}

func joinLines(lines []diff.Line) string {
	parts := make([]string, len(lines))
	for i, l := range lines {
		parts[i] = l.Text
	}
	return strings.Join(parts, "\n")
}
