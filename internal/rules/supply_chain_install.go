package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	scInstallScript = regexp.MustCompile(`"(?:postinstall|preinstall|install)"\s*:`)
	scScriptsOn     = regexp.MustCompile(`(?i)^\s*ignore-scripts\s*=\s*"?(?:false|0)"?\s*$`)
)

func installScriptFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) || !scInstallScript.MatchString(l.Text) {
			continue
		}
		out = add(out, scInput(f, l.Line, finding.InstallScript, "Install-time script", "CWE-829", finding.High))
	}
	return out
}

// isNpmrcComment also covers the ini `;` comment style.
func isNpmrcComment(text string) bool {
	return IsCommentLine(text) || strings.HasPrefix(strings.TrimSpace(text), ";")
}

// npmrcFindings flags registry hosts outside the allowlist and npmrc files
// that turn lifecycle scripts back on.
func npmrcFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		if isNpmrcComment(l.Text) {
			continue
		}
		if scScriptsOn.MatchString(l.Text) && !isTestPath(f.Path) {
			out = add(out, scInput(f, l.Line, finding.InstallScript, "Lifecycle scripts re-enabled", "CWE-829", finding.High))
			continue
		}
		m := scRegistryLine.FindStringSubmatch(l.Text)
		if m == nil || registryHostAllowed(m[1]) {
			continue
		}
		out = add(out, scInput(f, l.Line, finding.RegistryHost, "Unexpected npm registry host", "CWE-829", finding.High))
	}
	return out
}
