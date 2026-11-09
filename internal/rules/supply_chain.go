package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	scFirstParty    = regexp.MustCompile(`^(?:actions|github)/`)
	scPinnedSHA     = regexp.MustCompile(`(?i)@[0-9a-f]{40}$`)
	scNpmHost       = regexp.MustCompile(`registry\.npmjs\.org`)
	scInstallScript = regexp.MustCompile(`"(?:postinstall|preinstall|install)"\s*:`)
	scRegistryLine  = regexp.MustCompile(`^(?:registry|@[\w-]+:registry)\s*=\s*(\S+)`)
	scWorkflowFile  = regexp.MustCompile(`\.ya?ml$`)
	scUses          = regexp.MustCompile(`\buses:\s*(\S+)`)
	scNpmInstall    = regexp.MustCompile(`\brun:\s*npm\s+install\b`)
	scFrozen        = regexp.MustCompile(`--frozen-lockfile`)
	scLockfile      = regexp.MustCompile(`(?:package-lock\.json|yarn\.lock|pnpm-lock\.yaml)$`)
	scResolved      = regexp.MustCompile(`"resolved"\s*:\s*"([^"]+)"`)
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

	if name == "package.json" {
		for _, l := range f.Added {
			if IsCommentLine(l.Text) || !scInstallScript.MatchString(l.Text) {
				continue
			}
			out = add(out, scInput(f, l.Line, finding.InstallScript, "Install-time script", "CWE-829", finding.High))
		}
	}

	if strings.HasSuffix(name, ".npmrc") {
		for _, l := range f.Added {
			if IsCommentLine(l.Text) {
				continue
			}
			m := scRegistryLine.FindStringSubmatch(l.Text)
			if m == nil || scNpmHost.MatchString(m[1]) {
				continue
			}
			out = add(out, scInput(f, l.Line, finding.RegistryHost, "Unexpected npm registry host", "CWE-829", finding.High))
		}
	}

	if strings.Contains(f.Path, ".github/workflows/") && scWorkflowFile.MatchString(f.Path) {
		out = append(out, workflowFindings(f)...)
	}

	if scLockfile.MatchString(f.Path) {
		for _, l := range f.Added {
			m := scResolved.FindStringSubmatch(l.Text)
			if m != nil && !scNpmHost.MatchString(m[1]) {
				out = add(out, scInput(f, l.Line, finding.RegistryHost, "Lockfile resolved to unexpected registry", "CWE-829", finding.High))
			}
		}
	}
	return out
}

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
		if scNpmInstall.MatchString(l.Text) && !scFrozen.MatchString(l.Text) {
			out = add(out, scInput(f, l.Line, finding.LockfileRewrite, "CI install can rewrite the lockfile", "CWE-1104", finding.Medium))
		}
	}
	return out
}
