package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	// A secret-named variable carrying a framework public prefix is bundled
	// into the browser by construction.
	publicSecretAssign = regexp.MustCompile(`(?i)\b(?:NEXT_PUBLIC|NUXT_PUBLIC|VITE|PUBLIC|REACT_APP|EXPO_PUBLIC)_\w*(?:SECRET|CLOUDFLARE_API_TOKEN|CF_API_TOKEN|API_TOKEN)\w*["']?\s*[=:]\s*(.+?)\s*$`)
	// Turnstile secret in client code: an identifier naming it, assigned a literal of its shape.
	clientSecretAssign = regexp.MustCompile(`(?i)\b\w*(?:turnstile\w*secret|cloudflare_?api_?token|cf_?api_?token)\w*["']?\s*[=:]\s*["'](0x4[\w-]{20,}|[\w-]{40})["']`)
	clientExt          = regexp.MustCompile(`(?i)\.(?:tsx|jsx|vue|svelte|astro|html)$`)
	clientDir          = regexp.MustCompile(`(?i)(?:^|/)(?:client|public|static|components|islands)/`)
)

// publicEnvSecretFindings flags a Turnstile secret key or Cloudflare API token
// that reaches the browser: a public-prefixed variable holding a literal, or a
// literal assigned in client code. Placeholders and Turnstile test keys stay silent.
func publicEnvSecretFindings(f diff.File) []finding.Finding {
	client := clientExt.MatchString(f.Path) || clientDir.MatchString(f.Path) || strings.HasSuffix(baseLower(f.Path), "_headers")
	var out []finding.Finding
	for _, l := range f.Added {
		if IsCommentLine(l.Text) {
			continue
		}
		if m := publicSecretAssign.FindStringSubmatch(l.Text); m != nil && isLiteralSecretValue(publicValue(m[1])) {
			out = add(out, cfSecret(f, l.Line, "Secret exposed through a public env variable"))
			continue
		}
		if m := clientSecretAssign.FindStringSubmatch(l.Text); client && m != nil && isLiteralSecretValue(m[1]) {
			out = add(out, cfSecret(f, l.Line, "Cloudflare secret in client code"))
		}
	}
	return out
}

// publicValue unwraps an env or object value: quotes, trailing comma or semicolon, comment.
func publicValue(v string) string {
	v = stripLineComment(v)
	v = strings.TrimRight(v, ",;")
	if strings.HasPrefix(v, "process.") || strings.HasPrefix(v, "import.meta") || strings.HasPrefix(v, "env.") {
		return ""
	}
	return strings.Trim(strings.TrimSpace(v), `"'`)
}
