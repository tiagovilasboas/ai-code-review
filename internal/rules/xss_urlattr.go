package rules

import "regexp"

var (
	// xssHrefValue captures a dotted member path bound to an href (JSX or Vue).
	xssHrefValue = regexp.MustCompile(`(?:\bhref=\{|(?::|v-bind:)href=")\s*([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)+)\s*(?:\}|")`)
	// xssURLish limits the check to properties that name a URL.
	xssURLish = regexp.MustCompile(`(?i)(?:website|url|uri|link|homepage|site|href)$`)
	// xssURLCheck recognizes scheme or validity checks on a URL value.
	xssURLCheck = regexp.MustCompile(`/\^https?:|startsWith\(\s*['"]https?:|\bnew URL\(|\.protocol\b|\b(?:is|assert|validate|sanitize|safe|check)\w*(?:Url|URL|Uri|Link)\w*\s*\(`)
)

// unsafeURLAttr reports an href bound to a URL-named member path (a
// javascript: URL vector) when no scheme check appears in the hunk.
func (c xssContext) unsafeURLAttr(text string) bool {
	if c.urlChecked {
		return false
	}
	m := xssHrefValue.FindStringSubmatch(text)
	return m != nil && xssURLish.MatchString(m[1])
}
