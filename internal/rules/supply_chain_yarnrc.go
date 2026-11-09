package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var scYarnHosts = map[string]bool{
	"registry.yarnpkg.com": true,
	"registry.npmjs.org":   true,
}

// Top-level and scoped (npmScopes.*) registry keys, plus proxy settings.
var scYarnrcKey = regexp.MustCompile(`^\s*(npmRegistryServer|httpProxy|httpsProxy)\s*:\s*(.*)$`)

// yarnrcFindings flags Yarn Berry registry and proxy settings that point at a
// host outside the official registries.
func yarnrcFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) {
			continue
		}
		m := scYarnrcKey.FindStringSubmatch(l.Text)
		if m == nil {
			continue
		}
		value := m[2]
		if i := strings.Index(value, " #"); i >= 0 {
			value = value[:i]
		}
		if strings.TrimSpace(value) == "" || hostAllowedIn(scYarnHosts, value) {
			continue
		}
		out = add(out, scInput(f, l.Line, finding.RegistryHost, "Unexpected Yarn registry or proxy host", "CWE-829", finding.High))
	}
	return out
}
