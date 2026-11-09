package rules

import "regexp"

var (
	// ssrfFixedHost: the URL starts with a literal scheme and a complete host,
	// so user input can only reach the path or query.
	ssrfFixedHost = regexp.MustCompile("^[`'\"]https?://[A-Za-z0-9.-]+[:/?#'\"`]")
	// ssrfConfigBase: the URL starts from deployment configuration.
	ssrfConfigBase = regexp.MustCompile("^(?:`\\$\\{\\s*)?(?:process\\.env|import\\.meta\\.env)\\.\\w+")
	// ssrfConstBase: the URL starts from an ALL_CAPS constant.
	ssrfConstBase = regexp.MustCompile(`^[A-Z][A-Z0-9_]*\s*(?:\+|$)`)
	// ssrfAllowCheck recognizes a host or URL allowlist check earlier in the hunk.
	ssrfAllowCheck = regexp.MustCompile(`(?i)\b\w*allow\w*\.(?:has|includes|some|test|indexOf)\(|\b(?:allow(?:ed)?_?(?:hosts?|urls?|domains?|origins?)|allowlist|whitelist)\b|\b(?:assert|validate|check|verify|ensure)\w*(?:url|host|origin)\w*\s*\(|\bis(?:allowed|safe|public|trusted)\w*\s*\(`)
)

// ssrfFixedTarget reports whether the destination expression cannot be steered
// by a request: a constant, a fixed host, or a configured base URL.
func ssrfFixedTarget(arg string) bool {
	return isPureLiteral(arg) || ssrfFixedHost.MatchString(arg) ||
		ssrfConfigBase.MatchString(arg) || ssrfConstBase.MatchString(arg)
}
