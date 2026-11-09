package rules

import (
	"regexp"
	"strings"
)

var (
	// Key/value DSN password: "host=db user=app password=..." or "...;Password=...;".
	dsnKVPassword = regexp.MustCompile("(?i)\\b(?:password|pwd)=([^\\s\"'`;]+)")
	// Other connection keys that mark the string as a DSN, not a stray password=x.
	dsnKVContext = regexp.MustCompile(`(?i)\b(?:host|hostaddr|server|data source|dbname|database|initial catalog|user|user id|uid|sslmode|port)=`)
	dsnKVLocal   = regexp.MustCompile(`(?i)\b(?:host|server)=(?:localhost|127\.0\.0\.1|0\.0\.0\.0|\[::1\])\b`)
	// go-sql-driver/mysql DSN: user:password@tcp(host:3306)/db.
	dsnTCP = regexp.MustCompile(`([\w.\-]+):([^\s:@/"'()]+)@(?:tcp|unix)\(([^)]*)\)`)
)

// hasDSNPassword reports a connection string that embeds a literal password in
// key/value or driver (user:pass@tcp(host)) form. Placeholders, format verbs,
// env-built values and local development hosts stay silent.
func hasDSNPassword(text string) bool {
	if m := dsnKVPassword.FindStringSubmatch(text); m != nil &&
		dsnKVContext.MatchString(text) && !dsnKVLocal.MatchString(text) && isDSNSecret(m[1]) {
		return true
	}
	for _, m := range dsnTCP.FindAllStringSubmatch(text, -1) {
		host := strings.ToLower(m[3])
		if i := strings.LastIndex(host, ":"); i >= 0 {
			host = host[:i]
		}
		if !localHosts[host] && m[1] != m[2] && isDSNSecret(m[2]) {
			return true
		}
	}
	return false
}

func isDSNSecret(pass string) bool {
	if isPlaceholderValue(pass) || strings.ContainsAny(pass, "%{}") {
		return false
	}
	switch strings.ToLower(pass) {
	case "password", "pass", "passwd", "secret", "root", "admin":
		return false
	}
	return true
}
