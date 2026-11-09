package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	wfExpression = regexp.MustCompile(`\$\{\{\s*([^}]*?)\s*\}\}`)
	// wfAttackerField is a context an outside user writes: issue, pull request,
	// comment and review text, branch names, commit messages and author names.
	wfAttackerField = regexp.MustCompile(`(?i)\bgithub\.head_ref\b|\bgithub\.event\.(?:` +
		`issue\.(?:title|body)|` +
		`pull_request\.(?:title|body|head\.(?:ref|label|repo\.default_branch))|` +
		`(?:comment|review|review_comment)\.body|` +
		`discussion\.(?:title|body)|` +
		`pages(?:\[[^\]]*\]|\.\*)\.page_name|` +
		`commits(?:\[[^\]]*\]|\.\*)\.(?:message|author\.(?:email|name))|` +
		`head_commit\.(?:message|author\.(?:email|name))|` +
		`workflow_run\.(?:head_branch|head_commit\.message)` +
		`)\b`)
	// wfBooleanExpr marks expressions that yield true or false, not the field text.
	wfBooleanExpr = regexp.MustCompile(`==|!=|\b(?:contains|startsWith|endsWith)\s*\(`)
)

func isWorkflowFile(p string) bool {
	return strings.Contains(p, ".github/workflows/") && scWorkflowFile.MatchString(p)
}

// workflowCmdInjection flags GitHub Actions run steps that expand an
// attacker-controlled event field into the shell script (CWE-78). Passing the
// field through env: and reading it as a shell variable is the safe form.
func workflowCmdInjection(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, r := range addedRunCommands(f) {
		if !expandsAttackerField(r.cmd) {
			continue
		}
		out = add(out, finding.Input{
			Title:    "Workflow run step expands attacker-controlled event field",
			Class:    finding.CmdInjection,
			Path:     f.Path,
			Line:     r.line,
			CWE:      "CWE-78",
			OWASP:    finding.LabelsFor(finding.CmdInjection),
			Severity: finding.High,
		})
	}
	return out
}

func expandsAttackerField(cmd string) bool {
	for _, m := range wfExpression.FindAllStringSubmatch(cmd, -1) {
		if wfAttackerField.MatchString(m[1]) && !wfBooleanExpr.MatchString(m[1]) {
			return true
		}
	}
	return false
}
