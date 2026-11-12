package rules

import (
	"path"
	"regexp"
	"strings"
)

// phpConfigCredential is a PHP config-array entry: 'smtp_password' => 'literal'.
// The key must end in a credential noun; group 1 is the key, group 2 the value.
var phpConfigCredential = regexp.MustCompile(`(?i)["']([\w.-]*?(?:password|passwd|secret|token|api[_-]?key|private[_-]?key|access[_-]?key|credentials?))["']\s*=>\s*["']([^"']*)["']`)

var (
	// phpMask is a display mask ('**********', '••••••••'), not a credential.
	phpMask = regexp.MustCompile(`^[*xX#•.\-_]+$`)
	// phpValidationRule is a Laravel rule string: 'required|string|max:1000'.
	phpValidationRule = regexp.MustCompile(`^[a-z_]+(?::[\w,.\-]+)?(?:\|[a-z_]+(?::[\w,.\-]+)?)*$`)
)

// hasPHPConfigCredential reports a Laravel/Laminas/CodeIgniter config array
// entry that holds a live-looking credential literal. env('X'), getenv('X'),
// empty values, placeholders, weak dev words, prose labels, translation keys
// plain words (no digit or symbol), display masks and validation rule strings
// stay silent.
func hasPHPConfigCredential(p, text string) bool {
	if strings.ToLower(path.Ext(p)) != ".php" {
		return false
	}
	for _, m := range phpConfigCredential.FindAllStringSubmatch(text, -1) {
		v := m[2]
		if len(v) >= 8 && !strings.Contains(v, " ") && isLiveSecretValue(v) &&
			!weakDevValues[strings.ToLower(v)] && hasDigitOrSymbol(v) && !secretRefName.MatchString(v) &&
			!phpMask.MatchString(v) && !phpValidationRule.MatchString(v) {
			return true
		}
	}
	return false
}

// hasDigitOrSymbol reports a digit or a punctuation character other than the
// separators of a plain identifier (. - _ /).
func hasDigitOrSymbol(v string) bool {
	for _, r := range v {
		if r >= '0' && r <= '9' || strings.ContainsRune("!@#$%^&*()+=?~|;:,<>{}[]", r) {
			return true
		}
	}
	return false
}
