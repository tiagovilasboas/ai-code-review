package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// Cloudflare flags config-as-code that leaks a secret to the bundle or the
// browser, or turns an edge protection down: wrangler config, Terraform for
// the cloudflare provider, public env files and Worker code. It reads the
// diff only; it never calls Cloudflare. Test paths stay silent.
func Cloudflare(f diff.File) []finding.Finding {
	if isTestPath(f.Path) {
		return nil
	}
	out := wranglerFindings(f)
	out = append(out, publicEnvSecretFindings(f)...)
	out = append(out, terraformSettingFindings(f)...)
	out = append(out, terraformRulesetFindings(f)...)
	return append(out, accessHeaderFindings(f)...)
}

// turnstileTestSecret matches Cloudflare's documented Turnstile dummy secret keys.
var turnstileTestSecret = regexp.MustCompile(`(?i)^[123]x0{20,}[a-z]{2}$`)

// cfConfig builds a cloudflare-config finding; extra are additional official OWASP IDs.
func cfConfig(f diff.File, line int, title, cwe string, sev finding.Severity, extra ...string) finding.Input {
	return finding.Input{
		Title:    title,
		Class:    finding.CloudflareConfig,
		Path:     f.Path,
		Line:     line,
		CWE:      cwe,
		OWASP:    append(finding.LabelsFor(finding.CloudflareConfig), extra...),
		Severity: sev,
	}
}

// cfSecret builds a hardcoded-credential finding for a Cloudflare setting.
func cfSecret(f diff.File, line int, title string) finding.Input {
	return finding.Input{
		Title:    title,
		Class:    finding.Secret,
		Path:     f.Path,
		Line:     line,
		CWE:      "CWE-798",
		OWASP:    finding.LabelsFor(finding.Secret),
		Severity: finding.Critical,
	}
}

func baseLower(p string) string { return strings.ToLower(path.Base(p)) }

// isLiteralSecretValue reports a value that looks like a live credential:
// not empty, a placeholder, a reference, a label or a Turnstile test secret.
func isLiteralSecretValue(v string) bool {
	v = strings.Trim(strings.TrimSpace(v), `"'`)
	return v != "" && isLiveSecretValue(v) &&
		!weakDevValues[strings.ToLower(v)]
}

// stripLineComment drops a trailing # comment from a TOML or env value.
func stripLineComment(v string) string {
	return strings.TrimSpace(inlineComment.ReplaceAllString(v, ""))
}
