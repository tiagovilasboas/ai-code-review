package rules

import "strings"

// hostOfURL extracts the lowercase host of a URL-ish value without relying on
// url.Parse, so ${VAR} placeholders in userinfo do not make parsing fail. The
// userinfo is dropped (the host is whatever follows the last "@" of the
// authority), so "pypi.org@evil.example" resolves to evil.example.
func hostOfURL(raw string) string {
	raw = strings.Trim(strings.TrimSpace(raw), `"'`)
	if i := strings.Index(raw, "://"); i >= 0 {
		raw = raw[i+3:]
	} else {
		raw = strings.TrimPrefix(raw, "//")
	}
	if i := strings.IndexAny(raw, "/?#"); i >= 0 {
		raw = raw[:i]
	}
	if i := strings.LastIndex(raw, "@"); i >= 0 {
		raw = raw[i+1:]
	}
	if i := strings.LastIndex(raw, ":"); i >= 0 && !strings.HasSuffix(raw, "]") {
		raw = raw[:i]
	}
	return strings.TrimSuffix(strings.ToLower(raw), ".")
}

// hostAllowedIn compares the host exactly against an allowlist. A value that
// is an unresolved placeholder (${VAR}, $VAR, %VAR%) is not provably foreign
// and stays silent; an empty host is never allowed.
func hostAllowedIn(allowed map[string]bool, raw string) bool {
	raw = strings.Trim(strings.TrimSpace(raw), `"'`)
	if strings.HasPrefix(raw, "${") || strings.HasPrefix(raw, "$") || strings.HasPrefix(raw, "%") {
		return true
	}
	host := hostOfURL(raw)
	return host != "" && allowed[host]
}
