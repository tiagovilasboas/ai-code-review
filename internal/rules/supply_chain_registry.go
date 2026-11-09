package rules

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// Hosts compared exactly, never by substring.
var scRegistryHosts = map[string]bool{
	"registry.npmjs.org":   true,
	"registry.yarnpkg.com": true,
}

var (
	scRegistryLine = regexp.MustCompile(`^\s*(?:registry|@[\w.-]+:registry)\s*=\s*(\S+)`)
	scLockfile     = regexp.MustCompile(`(?:package-lock\.json|yarn\.lock|pnpm-lock\.yaml)$`)
	// A resolved/tarball source in package-lock.json, yarn.lock or pnpm-lock.yaml.
	scLockURL = regexp.MustCompile(`(?:"resolved"\s*:|^\s*resolved\b|\btarball\s*:)\s*"?((?:https?|git\+[a-z]+|git|ssh)://[^"\s,}#]+)`)
	// A package.json dependency whose spec is a URL instead of a version range.
	scDepURL = regexp.MustCompile(`^\s*"([^"]+)"\s*:\s*"((?:https?|git\+[a-z]+|git|ssh)://[^"]+)"`)
)

// Dependency-spec keys that hold metadata URLs, not install sources.
var scMetadataKeys = map[string]bool{
	"homepage": true, "url": true, "repository": true, "bugs": true, "funding": true,
	"author": true, "license": true, "main": true, "types": true, "email": true,
}

// registryHostAllowed parses the host of a registry or tarball URL and
// compares it exactly. Unparseable values are not allowed; an unresolved
// ${VAR} placeholder is not provably foreign and stays silent.
func registryHostAllowed(raw string) bool {
	raw = strings.Trim(strings.TrimSpace(raw), `"'`)
	if strings.HasPrefix(raw, "${") {
		return true
	}
	if !strings.Contains(raw, "://") && !strings.HasPrefix(raw, "//") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	return scRegistryHosts[host]
}

func lockfileRegistryFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) {
			continue
		}
		m := scLockURL.FindStringSubmatch(l.Text)
		if m != nil && !registryHostAllowed(m[1]) {
			out = add(out, scInput(f, l.Line, finding.RegistryHost, "Lockfile resolved to unexpected registry", "CWE-829", finding.High))
		}
	}
	return out
}

func dependencyURLFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, l := range f.Added {
		m := scDepURL.FindStringSubmatch(l.Text)
		if m == nil || scMetadataKeys[strings.ToLower(m[1])] || registryHostAllowed(m[2]) {
			continue
		}
		out = add(out, scInput(f, l.Line, finding.RegistryHost, "Dependency installed from a URL outside the registry", "CWE-829", finding.High))
	}
	return out
}
