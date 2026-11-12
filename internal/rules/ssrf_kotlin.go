package rules

import (
	"regexp"
	"strings"
)

var (
	// ssrfKotlinCall is a Ktor HttpClient request: client.get(url), client.post(url) { ... },
	// client.request(url). The destination is the first argument.
	ssrfKotlinCall = regexp.MustCompile(`\b\w*(?i:client|http)\w*\.(get|post|put|patch|delete|head|options|request|submitForm|prepareGet|preparePost)\s*\(\s*(.*)$`)
	// ktNamedURL drops a named destination argument: get(urlString = target).
	ktNamedURL = regexp.MustCompile(`^(?:urlString|url)\s*=\s*`)
	// ktTemplate is a Kotlin string template: $name or ${expr}.
	ktTemplate = regexp.MustCompile(`\$\{([^}]*)\}|\$([A-Za-z_]\w*)`)
)

// ssrfKotlinDest returns the destination of a Ktor HttpClient call on text.
// String templates become concatenations, so the shared scanner sees the
// values they interpolate.
func ssrfKotlinDest(text string) (string, bool) {
	m := ssrfKotlinCall.FindStringSubmatch(text)
	if m == nil {
		return "", false
	}
	args := splitCallArgs(m[2])
	if len(args) == 0 {
		return "", false
	}
	return ktExpandTemplates(ktNamedURL.ReplaceAllString(args[0], "")), true
}

// ktExpandTemplates rewrites "a$b${c}d" as "a" + b + c + "d".
func ktExpandTemplates(s string) string {
	if !strings.Contains(s, "$") {
		return s
	}
	return ktTemplate.ReplaceAllStringFunc(s, func(t string) string {
		return `" + ` + strings.Trim(t, "${}") + ` + "`
	})
}
