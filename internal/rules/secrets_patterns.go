package rules

import (
	"regexp"
	"strings"
)

var (
	// Connection URL with credentials: scheme://user:password@host.
	secretDBURL = regexp.MustCompile(`(?i)\b(?:postgres(?:ql)?|mysql|mariadb|mongodb(?:\+srv)?|rediss?|amqps?|mssql|sqlserver)://([^\s:@/"']*):([^\s@/"']+)@([^\s/:"'?]*)`)
	// PEM private key header (public keys and certificates do not match).
	secretPEM = regexp.MustCompile(`-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY(?: BLOCK)?-----`)
	// .npmrc-style registry credential: //host/:_authToken=value, _auth=value, _password=value.
	secretRC = regexp.MustCompile(`^\s*(?://[^\s=]*:)?_(?:authToken|auth|password)\s*=\s*(\S+)`)

	placeholderWords = []string{"example", "dummy", "fake", "placeholder", "changeme", "change-me", "not-a-real", "your-", "your_", "xxxx", "redacted", "<", "${", "{{", "%s"}
	// Names of cookies, headers or storage keys: kebab/snake words that end in a naming noun.
	secretNameShape = regexp.MustCompile(`^(?:x-[a-z0-9-]+|[a-z]+(?:[-_.][a-z]+)*[-_.](?:token|cookie|name|id|header))$`)
	testPathShape   = regexp.MustCompile(`(?i)(?:^|/)(?:tests?|__tests__|__mocks__|fixtures?|e2e|spec)/|\.(?:test|spec)\.[a-z]+$`)
	localHosts      = map[string]bool{"localhost": true, "127.0.0.1": true, "0.0.0.0": true, "[::1]": true}
)

// isPlaceholderValue reports whether a literal is obviously not a live secret.
func isPlaceholderValue(v string) bool {
	lv := strings.ToLower(strings.TrimSpace(v))
	if lv == "" || strings.HasPrefix(lv, "$") {
		return true
	}
	for _, w := range placeholderWords {
		if strings.Contains(lv, w) {
			return true
		}
	}
	return false
}

// isNameLiteral reports whether a literal names a cookie, header or key slot.
func isNameLiteral(v string) bool { return secretNameShape.MatchString(v) }

func isTestPath(path string) bool { return testPathShape.MatchString(path) }

// hasLiteralURLPassword reports a credentialed connection URL whose password is
// a real-looking literal: not interpolated, not a user==password pair, not local.
func hasLiteralURLPassword(text string) bool {
	for _, m := range secretDBURL.FindAllStringSubmatch(text, -1) {
		user, pass, host := strings.ToLower(m[1]), strings.ToLower(m[2]), strings.ToLower(m[3])
		if localHosts[host] || user == pass || isPlaceholderValue(pass) {
			continue
		}
		switch pass {
		case "password", "pass", "passwd", "secret", "root", "admin":
			continue
		}
		return true
	}
	return false
}

// hasRCCredential reports a literal registry credential in .npmrc syntax.
func hasRCCredential(text string) bool {
	m := secretRC.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	v := strings.Trim(m[1], `"'`)
	return v != "" && !strings.HasPrefix(v, "$") && !strings.Contains(v, "${")
}
