package rules

import (
	"regexp"
	"strings"
)

var addressee = regexp.MustCompile(`(?i)\b(?:ai|a\.i\.|llms?|assistants?|automated\s+(?:code\s+)?(?:review|scan|analy|tool|check|system|agent|bot)\w*|(?:review|reviewer|pr)\s+(?:bot|agent)s?|copilot|claude|chatgpt|gpt|language\s+models?)\b`)

// bareVerb is the weakest directive: "merge" and "approve" are everyday words
// in any text about pull requests.
var bareVerb = regexp.MustCompile(`(?i)\b(?:approve|merge)\b`)

// A bare verb counts only when aimed at an addressee: anywhere after a
// vocative ("assistant: ... safe to merge"), up to bareVerbAfter bytes after a
// plain mention, or just before one across a comma ("merge it, Claude").
const (
	bareVerbAfter  = 40
	bareVerbBefore = 16
)

// vocative matches what follows an addressee that opens a direct address:
// "assistant:", "AI reviewers:", "Automated review agents:".
var vocative = regexp.MustCompile(`^(?:\s+\w+){0,2}\s*:`)

// addresseeSpans returns the addressee matches that are words in prose. A
// match inside a path, file name or identifier (ai-code-review, .claude/,
// CLAUDE.md) names a thing, not a reader.
func addresseeSpans(text string) [][]int {
	var out [][]int
	for _, m := range addressee.FindAllStringIndex(text, -1) {
		if !identifierPart(text, m[0], m[1]) {
			out = append(out, m)
		}
	}
	return out
}

func identifierPart(text string, start, end int) bool {
	if start > 0 && strings.IndexByte("._@", text[start-1]) >= 0 {
		return true
	}
	if end < len(text) && (text[end] == '_' || text[end] == '.' && end+1 < len(text) && isWordByte(text[end+1])) {
		return true
	}
	// A slash anywhere in the surrounding token makes it a path or URL.
	lo := strings.LastIndexAny(text[:start], " \t") + 1
	hi := len(text)
	if i := strings.IndexAny(text[end:], " \t"); i >= 0 {
		hi = end + i
	}
	return strings.Contains(text[lo:hi], "/")
}

// addressedBareVerb reports whether a bare approve or merge is aimed at an
// addressee in the same sentence.
func addressedBareVerb(text string, spans [][]int) bool {
	for _, v := range bareVerb.FindAllStringIndex(text, -1) {
		for _, a := range spans {
			if a[1] <= v[0] && (vocative.MatchString(text[a[1]:]) || sameSentence(text[a[1]:v[0]], bareVerbAfter)) {
				return true
			}
			if v[1] <= a[0] && sameSentence(text[v[1]:a[0]], bareVerbBefore) && strings.Contains(text[v[1]:a[0]], ",") {
				return true
			}
		}
	}
	return false
}

func sameSentence(gap string, limit int) bool {
	return len(gap) <= limit && !strings.ContainsAny(gap, ".;!?")
}

// docQuotes match text a document quotes: straight or curly double quotes and
// code spans.
var docQuotes = regexp.MustCompile("\"[^\"]*\"|\u201c[^\u201d]*\u201d|`[^`]*`")

// quotedMention reports whether every instruction-shaped part of a docs line
// sits inside quotation marks: the document talks about the phrase instead of
// saying it.
func quotedMention(text string) bool {
	return !instructionShaped(docQuotes.ReplaceAllString(text, " "))
}
