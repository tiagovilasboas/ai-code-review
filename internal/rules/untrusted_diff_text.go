package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// directiveAgent phrases can only be addressed to a model that reads the diff.
const directiveAgentSrc = `ignore\s+(?:all\s+|any\s+)?(?:your\s+|the\s+)?(?:previous|prior|above|earlier)\s+(?:instructions|rules|guidelines|prompts?)` +
	`|ignore\s+your\s+(?:rules|instructions|guidelines)` +
	`|disregard\s+(?:the\s+|all\s+|any\s+)?(?:guardrails?|evidence|instructions|rules|above|previous)` +
	`|system\s+prompt\s+override`

// directiveStrong phrases are aimed at the reviewer by construction. The
// merge/approve orders also read as UI copy ("You must approve this PR first").
var (
	directiveAgent  = regexp.MustCompile(`(?i)(?:` + directiveAgentSrc + `)`)
	directiveStrong = regexp.MustCompile(`(?i)(?:` + directiveAgentSrc +
		`|you\s+must\s+(?:merge|approve)` +
		`|(?:merge|approve)\s+this\s+(?:pr|pull request|diff))`)
)

// directiveWeak phrases are ordinary English in app code (review apps, Sentry
// reporting); they count only when the same text names an AI addressee. The
// bare verbs approve and merge are weaker still: see addressedBareVerb.
var directiveWeak = regexp.MustCompile(`(?i)(?:` +
	`skip\s+(?:this\s+|the\s+|all\s+)?(?:review|checks?)` +
	`|(?:do\s+not|don'?t)\s+(?:report|flag|raise)|not\s+raise` +
	`|report\s+(?:no|nothing|zero)|no\s+(?:findings|issues)|nothing\s+to\s+(?:flag|report)` +
	`|no\s+need\s+to\s+(?:look|review|check|inspect)|pre-?approved` +
	`|(?:mark|treat)\s+(?:this\s+)?(?:as\s+)?(?:safe|clean|approved)` +
	`|without\s+(?:further\s+)?(?:review|analysis))`)

var quoted = regexp.MustCompile("\"(?:[^\"\\\\]|\\\\.)*\"|'(?:[^'\\\\]|\\\\.)*'|`[^`]*`")

var docExts = map[string]bool{".md": true, ".mdx": true, ".markdown": true, ".txt": true, ".rst": true, ".adoc": true}

func instructionShaped(text string) bool {
	if directiveStrong.MatchString(text) {
		return true
	}
	spans := addresseeSpans(text)
	return len(spans) > 0 && (directiveWeak.MatchString(text) || addressedBareVerb(text, spans))
}

type blockDelim struct{ open, close string }

var blockDelims = []blockDelim{{"/*", "*/"}, {"<!--", "-->"}, {`"""`, `"""`}, {"'''", "'''"}}

// advanceBlock tracks multi-line comments and docstrings. It returns the
// closer still pending after the line and whether the line sat in or opened a block.
func advanceBlock(closer, text string) (string, bool) {
	inside := closer != ""
	if closer != "" {
		if strings.Contains(text, closer) {
			return "", true
		}
		return closer, true
	}
	for _, d := range blockDelims {
		i := strings.Index(text, d.open)
		if i < 0 {
			continue
		}
		if strings.Contains(text[i+len(d.open):], d.close) {
			return "", true
		}
		return d.close, true
	}
	return "", inside
}

// trailingComment returns the text after an end-of-line comment marker.
func trailingComment(text string) string {
	for _, m := range []string{" // ", " #"} {
		if i := strings.Index(text, m); i >= 0 {
			return text[i+len(m):]
		}
	}
	return ""
}

// candidates returns the parts of an added line a reader could take as
// instructions: comments, docs prose, and (outside tests) string literals.
func candidates(f diff.File, l diff.Line, inBlock bool) string {
	switch {
	case inBlock || IsCommentLine(l.Text):
		return l.Text
	case docExts[strings.ToLower(path.Ext(f.Path))]:
		return l.Text
	}
	parts := []string{trailingComment(l.Text)}
	if !isTestPath(f.Path) {
		parts = append(parts, quoted.FindAllString(l.Text, -1)...)
	}
	return strings.Join(parts, " ")
}
