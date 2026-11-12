package rules

import (
	"regexp"
	"strings"
)

var (
	// dotenvPath is a dotenv file: .env, .env.production, .env.local, prod.env.
	dotenvPath = regexp.MustCompile(`(?i)(?:^|/)(?:\.env(?:\.[\w.-]+)?|[\w.-]+\.env)$`)
	// dotenvTemplate is a committed template that holds placeholders by design.
	dotenvTemplate = regexp.MustCompile(`(?i)\.(?:example|sample|template|dist|tpl|defaults?)$`)
	dotenvAssign   = regexp.MustCompile(`^\s*(?:export\s+)?[A-Za-z_][\w.]*\s*=\s*(.*?)\s*$`)
	// providerKeyValue is a whole unquoted value shaped like a provider key.
	providerKeyValue = regexp.MustCompile(`^(?:(?:pk|sk)_live_[A-Za-z0-9]{12,}|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{20,}|github_pat_\w{20,}|xox[baprs]-[\w-]{10,})`)
	// sendGridKey: SG.<22 chars>.<20 or more chars>.
	sendGridKey = regexp.MustCompile(`\bSG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{20,}`)
)

// hasSendGridKey reports a SendGrid API key in any file or quoting.
func hasSendGridKey(text string) bool {
	m := sendGridKey.FindString(text)
	return m != "" && !isPlaceholderValue(m)
}

// hasDotenvCredential reports a live secret in a dotenv file: a credential
// key with a literal value, or a value shaped like a provider key. Template
// files (.env.example) and ${VAR} references stay silent.
func hasDotenvCredential(p, text string) bool {
	if !dotenvPath.MatchString(p) || dotenvTemplate.MatchString(p) {
		return false
	}
	if configLineCredential(strings.TrimPrefix(strings.TrimSpace(text), "export ")) {
		return true
	}
	m := dotenvAssign.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	v := strings.Trim(inlineComment.ReplaceAllString(m[1], ""), `"'`)
	return providerKeyValue.MatchString(v) && !isPlaceholderValue(v)
}
