package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	// Spring wildcard shapes: @CrossOrigin(origins|originPatterns = "*") and
	// registry.allowedOrigins("*") / allowedOriginPatterns("*").
	corsSpringStar = regexp.MustCompile(`\b(?:origins|originPatterns)\s*=\s*(?:\{\s*)?"\*"|\.allowedOrigin(?:Pattern)?s\(\s*"\*"`)
	// A local assigned from the request Origin, later echoed into the header.
	corsOriginAssign = regexp.MustCompile(`\b([A-Za-z_]\w*)\s*:?=\s*(.+)`)
	// Evidence that the origin was checked against a list before it is echoed.
	corsAllowCheck = regexp.MustCompile(`(?i)\b(?:allow(?:ed|list)\w*|whitelist\w*|trusted\w*)\b|slices\.Contains|\.includes\(|\.has\(|\.indexOf\(|\bcontains\(`)
)

const corsAllowLookback = 8

// originVars returns locals that hold the request Origin.
func originVars(code []diff.Line) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, l := range code {
		m := corsOriginAssign.FindStringSubmatch(l.Text)
		if m != nil && corsReflectHeader.MatchString(m[2]) && !corsHeaderName.MatchString(l.Text) {
			out = append(out, regexp.MustCompile(`\b`+regexp.QuoteMeta(m[1])+`\b`))
		}
	}
	return out
}

// echoesRequestOrigin reports an Allow-Origin header line whose value is the
// request Origin, read inline or through a local, with no allowlist check in
// the few lines above it.
func echoesRequestOrigin(l diff.Line, code []diff.Line, vars []*regexp.Regexp) bool {
	if !corsHeaderName.MatchString(l.Text) {
		return false
	}
	echoed := corsReflectHeader.MatchString(l.Text)
	for _, v := range vars {
		echoed = echoed || v.MatchString(corsHeaderName.ReplaceAllString(l.Text, ""))
	}
	if !echoed {
		return false
	}
	for _, c := range code {
		if c.Line >= l.Line-corsAllowLookback && c.Line <= l.Line && corsAllowCheck.MatchString(corsHeaderName.ReplaceAllString(c.Text, "")) {
			return false
		}
	}
	return true
}
