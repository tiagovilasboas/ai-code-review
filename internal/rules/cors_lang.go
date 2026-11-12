package rules

import (
	"regexp"
	"strings"
)

var (
	// Wildcard origin set through a framework option rather than a header:
	// FastAPI/Flask-CORS (Python), Django cors-headers, gin-contrib/cors and
	// rs/cors (Go), Spring CorsConfiguration, Laravel config/cors.php.
	corsLangStar = regexp.MustCompile(`\ballow_origins\s*=\s*\[\s*["']\*["']` +
		`|\borigins\s*=\s*(?:\[\s*)?["']\*["']|["']origins["']\s*:\s*(?:\[\s*)?["']\*["']` +
		`|^\s*CORS_(?:ORIGIN_)?ALLOW_ALL(?:_ORIGINS)?\s*=\s*True\b` +
		`|\bAllow(?:ed)?Origins\s*:\s*\[\]string\{\s*"\*"|\bAllowAllOrigins\s*:\s*true\b` +
		`|\bAllowOriginFunc\s*:\s*func\([^)]*\)\s*bool\s*\{\s*return true\s*\}` +
		`|\.(?:addAllowedOrigin|addAllowedOriginPattern)\(\s*"\*"` +
		`|\.setAllowedOrigin(?:Pattern)?s\(\s*(?:List|Arrays|Set|Collections)\.\w+\(\s*"\*"` +
		`|['"]allowed_origins['"]\s*=>\s*\[\s*['"]\*['"]`)
	// Credentials switches of the same frameworks.
	corsLangCredentials = `\ballow_credentials\s*=\s*True\b|\bsupports_credentials\s*=\s*True\b|\bCORS_ALLOW_CREDENTIALS\s*=\s*True\b` +
		`|\bAllowCredentials\s*:\s*true\b|\bsetAllowCredentials\(\s*true\b|['"]supports_credentials['"]\s*=>\s*true\b` +
		`|access-control-allow-credentials["']?\]?\s*=\s*["']true`
	// The request Origin read through a PHP superglobal, Django META or Laravel.
	corsLangOrigin = `|\bHTTP_ORIGIN\b|->header\(\s*["']origin["']\s*\)`
	// An allowlist membership test in PHP or Python.
	corsLangAllow = `|\bin_array\(|\borigin\s+in\b`
)

var corsFlaskCredentials = regexp.MustCompile(`\bCORS\s*\([^)]*\bsupports_credentials\s*=\s*True\b`)

// corsFlaskDefaultOrigins reports Flask-CORS with credentials and no origins
// option: the extension then allows every origin.
func corsFlaskDefaultOrigins(text string) bool {
	return corsFlaskCredentials.MatchString(text) && !strings.Contains(text, "origins")
}
