package rules

import "regexp"

var (
	// UPPER_CASE setting ending in a credential noun, assigned a quoted literal:
	// SECRET_KEY = '...', AWS_SECRET_ACCESS_KEY = "...", const JWT_SECRET = "...".
	secretSettingAssign = regexp.MustCompile(`^\s*(?:export\s+|const\s+|var\s+|let\s+)*(?:[A-Z][A-Z0-9_]*_)?(?:SECRET|SECRET_KEY|PASSWORD|PASSWD|TOKEN|API_?KEY|PRIVATE_KEY|ACCESS_KEY|CREDENTIALS?)\s*:?=\s*["']([^"']+)["']`)
	// Quoted UPPER_CASE dict key (Django DATABASES): 'PASSWORD': '...'.
	secretSettingKey = regexp.MustCompile(`["'](?:[A-Z][A-Z0-9_]*_)?(?:SECRET|PASSWORD|PASSWD|TOKEN|API_?KEY|PRIVATE_KEY|ACCESS_KEY)["']\s*:\s*["']([^"']+)["']`)
)

// hasSettingCredential reports an UPPER_CASE setting or settings-dict entry
// holding a credential literal. Environment lookups never match: the value
// must start with a quote.
func hasSettingCredential(text string) bool {
	for _, re := range []*regexp.Regexp{secretSettingAssign, secretSettingKey} {
		if m := re.FindStringSubmatch(text); m != nil && isLiveSecretValue(m[1]) {
			return true
		}
	}
	return false
}

// isLiveSecretValue reports a literal that is not a placeholder, label, name or resource path.
func isLiveSecretValue(v string) bool {
	return !isPlaceholderValue(v) && !isNameLiteral(v) && !isProseLiteral(v) && !isSecretResourceName(v)
}
