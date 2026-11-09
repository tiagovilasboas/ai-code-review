package rules

import "regexp"

const jwtQuote = "[\"'`]"

var (
	// jwt.sign(payload, "literal") / jwt.verify(token, "literal").
	jwtCallLiteral = regexp.MustCompile(`(?i)\bjwt\w*\.(?:sign|verify)\s*\(\s*(?:\{[^}]*\}|[^,(){}])+,\s*` + jwtQuote + `([^"'` + "`" + `]+)` + jwtQuote)
	// crypto.createHmac("sha256", "literal").
	hmacCallLiteral = regexp.MustCompile(`\bcreateHmac\s*\(\s*` + jwtQuote + `[\w-]+` + jwtQuote + `\s*,\s*` + jwtQuote + `([^"'` + "`" + `]+)` + jwtQuote)
	// jose-style key bytes: new TextEncoder().encode("literal").
	encoderLiteral = regexp.MustCompile(`TextEncoder\(\)\.encode\(\s*` + jwtQuote + `([^"'` + "`" + `]+)` + jwtQuote + `\s*\)`)
	encoderContext = regexp.MustCompile(`(?i)secret|key|jwt|hmac|sign`)
	// const jwtKey = "literal", signingSecret: "literal", HMAC_KEY = "literal".
	jwtNameAssign = regexp.MustCompile(`(?i)\b\w*(?:jwt|hmac|signing)[_-]?(?:key|secret)\s*[:=]\s*` + jwtQuote + `([^"'` + "`" + `]+)` + jwtQuote)
)

// hasJWTSecret reports a signing secret passed or assigned as a string literal
// in JS/TS style code. Variables, config reads and env lookups never match: the
// secret argument must be a quoted literal.
func hasJWTSecret(text string) bool {
	for _, re := range []*regexp.Regexp{jwtCallLiteral, hmacCallLiteral, jwtNameAssign} {
		if m := re.FindStringSubmatch(text); m != nil && isLiveSecretValue(m[1]) {
			return true
		}
	}
	if m := encoderLiteral.FindStringSubmatch(text); m != nil && encoderContext.MatchString(text) {
		return isLiveSecretValue(m[1])
	}
	return false
}
