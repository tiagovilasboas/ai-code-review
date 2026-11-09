package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

const sanitizerNames = `(?:\w*[Pp]urify\.sanitize|sanitizeHtml|sanitizeHTML|escapeHtml|escapeHTML|sanitize)`

var (
	// xssSanitizerCall matches a call to a known HTML sanitizer or escaper.
	xssSanitizerCall = regexp.MustCompile(`(?:await\s+)?\b` + sanitizerNames + `\s*\(`)
	// xssSanitizedAssign captures a variable assigned the sanitizer's result.
	xssSanitizedAssign = regexp.MustCompile(`([A-Za-z_$][\w$.]*)\s*=\s*(?:useMemo\(\s*\(\)\s*=>\s*)?(?:await\s+)?\b` + sanitizerNames + `\s*\(`)
)

// xssContext holds facts gathered from the whole file hunk, not one line.
type xssContext struct {
	sanitized  map[string]bool // variables holding sanitizer output
	urlChecked bool            // a URL scheme/validity check appears in the hunk
}

func newXSSContext(f diff.File) xssContext {
	c := xssContext{sanitized: map[string]bool{}}
	for _, l := range f.Lines {
		if IsCommentLine(l.Text) {
			continue
		}
		if m := xssSanitizedAssign.FindStringSubmatch(l.Text); m != nil {
			c.sanitized[m[1]] = true
		}
		if xssURLCheck.MatchString(l.Text) {
			c.urlChecked = true
		}
	}
	return c
}

// flags reports whether an added line holds an XSS sink.
func (c xssContext) flags(text string) bool {
	return c.sinkHit(text) || jqueryHTMLWrite(text) || c.unsafeURLAttr(text)
}
