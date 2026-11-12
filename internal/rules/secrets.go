package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	secretAssign = regexp.MustCompile(`(?i)(?:password|passwd|secret|token|api[_-]?key|private[_-]?key|access[_-]?key|credential)\s*[:=]\s*["']([^"']+)["']`)
	// qualifiedKey is a field or constant used as a map key (finding.Secret:);
	// its value is a label, not a credential.
	qualifiedKey = regexp.MustCompile(`(?i)[\w.]+\.(?:password|passwd|secret|token|api[_-]?key|private[_-]?key|access[_-]?key|credential)\s*:`)
	knownPrefix  = regexp.MustCompile(`["'](pk_live_|pk_test_|sk_live_|sk_test_|AKIA|ghp_|github_pat_|xox[baprs]-)([^"']*)`)
)

// Secrets flags hardcoded credentials (CWE-798) and wildcard CORS origins (CWE-942)
// on added lines. Lines reading process.env and comments are ignored.
func Secrets(f diff.File) []finding.Finding {
	var out []finding.Finding
	for _, line := range f.Added {
		// A leading "//" is the host prefix of an .npmrc credential, not a comment.
		if (IsCommentLine(line.Text) && !hasRCCredential(line.Text)) || strings.Contains(line.Text, "process.env") {
			continue
		}
		if hasCredentialLiteral(f.Path, line.Text) {
			out = add(out, finding.Input{
				Title:    "Hardcoded credential in source",
				Class:    finding.Secret,
				Path:     f.Path,
				Line:     line.Line,
				CWE:      "CWE-798",
				OWASP:    finding.LabelsFor(finding.Secret),
				Severity: finding.Critical,
			})
		}
	}
	return append(out, corsFindings(f)...)
}

// minTestTokenTail is the shortest random tail that makes a token-shaped value
// in a test file look like a real key rather than a placeholder (sk_live_x).
const minTestTokenTail = 12

// hasKnownToken reports a string with a known credential prefix. In a test
// path a Stripe test-mode key (pk_test_, sk_test_) or a value with a tail under
// minTestTokenTail is a placeholder; a live-looking key still counts there.
func hasKnownToken(path, text string) bool {
	if hasSendGridKey(text) {
		return true
	}
	for _, m := range knownPrefix.FindAllStringSubmatch(text, -1) {
		if !isTestPath(path) {
			return true
		}
		if !strings.HasSuffix(m[1], "_test_") && len(m[2]) >= minTestTokenTail {
			return true
		}
	}
	return false
}

// hasCredentialLiteral reports a hardcoded credential on one added line.
// Known token prefixes, PEM private keys and credentialed URLs always count;
// generic name=literal assignments skip placeholders, naming constants, prose
// labels and test files.
func hasCredentialLiteral(path, text string) bool {
	if hasKnownToken(path, text) || secretPEM.MatchString(text) || hasLiteralURLPassword(text) || hasRCCredential(text) {
		return true
	}
	if isTestPath(path) {
		return false
	}
	if hasJWTSecret(text) || hasDSNPassword(text) || hasSettingCredential(text) || hasConfigCredential(path, text) || hasDotenvCredential(path, text) || hasPHPConfigCredential(path, text) {
		return true
	}
	for _, m := range secretAssign.FindAllStringSubmatch(qualifiedKey.ReplaceAllString(text, ""), -1) {
		if !isPlaceholderValue(m[1]) && !isNameLiteral(m[1]) && !isProseLiteral(m[1]) && !isSecretResourceName(m[1]) {
			return true
		}
	}
	return false
}
