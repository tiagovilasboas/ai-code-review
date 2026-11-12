package rules

import (
	"path"
	"regexp"
	"strings"
)

var (
	configFileExt = regexp.MustCompile(`(?i)\.(?:properties|ya?ml|toml|ini|cfg|conf)$`)
	// key.with.dots[=:]value where the last key segment ends in a credential noun.
	// Unlike code files, a qualified key is the whole point here (spring.datasource.password).
	configCredential = regexp.MustCompile(`(?i)^\s*(?:-\s+)?["']?[\w.\-\[\]]*?(password|passwd|secret|token|api[-_.]?key|private[-_.]?key|access[-_.]?key|credentials?)["']?\s*[=:]\s*(.*?)\s*$`)
	inlineComment    = regexp.MustCompile(`\s+#.*$`)
	// Lowercase words only: a reference to a secret by name (existingSecret: db-credentials).
	secretRefName = regexp.MustCompile(`^[a-z]+(?:[-_.][a-z]+)*$`)
	weakDevValues = map[string]bool{"password": true, "pass": true, "passwd": true, "secret": true, "root": true, "admin": true, "postgres": true, "test": true, "dev": true, "guest": true}
)

// hasConfigCredential reports a literal credential value in a properties, YAML
// or similar config file. ${VAR} placeholders, empty values, encrypted values,
// YAML tags/anchors and names of secrets stay silent.
func hasConfigCredential(p, text string) bool {
	return configFileExt.MatchString(path.Base(p)) && configLineCredential(text)
}

// configLineCredential reports a credential key assigned a live literal on one config line.
func configLineCredential(text string) bool {
	m := configCredential.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	v := strings.Trim(inlineComment.ReplaceAllString(m[2], ""), `"'`)
	if v == "" || strings.ContainsAny(v[:1], "!*&|>$%@{[<") || strings.HasPrefix(v, "ENC(") {
		return false
	}
	switch strings.ToLower(v) {
	case "true", "false", "null", "~":
		return false
	}
	lw := strings.ToLower(m[1])
	refKey := lw == "secret" || lw == "token" || strings.HasPrefix(lw, "credential")
	if refKey && secretRefName.MatchString(v) {
		return false
	}
	return isLiveSecretValue(v) && !weakDevValues[strings.ToLower(v)]
}
