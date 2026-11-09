package rules

import (
	"path"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func scInput(f diff.File, line int, class finding.Class, title, cwe string, sev finding.Severity) finding.Input {
	return finding.Input{
		Title: title, Class: class, Path: f.Path, Line: line,
		CWE: cwe, OWASP: finding.LabelsFor(class), Severity: sev,
	}
}

// SupplyChain flags install-time scripts, foreign registries, floating
// third-party Actions and CI installs that can rewrite the lockfile.
func SupplyChain(f diff.File) []finding.Finding {
	var out []finding.Finding
	name := path.Base(f.Path)

	if name == "package.json" && !isTestPath(f.Path) {
		out = append(out, installScriptFindings(f)...)
		out = append(out, dependencyURLFindings(f)...)
	}
	if strings.HasSuffix(name, ".npmrc") {
		out = append(out, npmrcFindings(f)...)
	}
	if strings.Contains(f.Path, ".github/workflows/") && scWorkflowFile.MatchString(f.Path) {
		out = append(out, workflowFindings(f)...)
	}
	if scLockfile.MatchString(f.Path) {
		out = append(out, lockfileRegistryFindings(f)...)
		out = append(out, lockfileIntegrityFindings(f)...)
	}
	if isPipFile(f.Path) {
		out = append(out, pipIndexFindings(f)...)
	}
	if name == ".yarnrc.yml" {
		out = append(out, yarnrcFindings(f)...)
	}
	return out
}
