package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	authzSink = regexp.MustCompile(`\b(?:findById|getById|findOne|findUnique|findFirst)\s*\(`)
	// Caller-controlled input: Express/Fastify/Koa request objects, or a route
	// context reached through ctx/context.
	authzCallerID = regexp.MustCompile(`\b(?:req|request)\.(?:params|query|body)\b|\b(?:ctx|context)\.(?:params|query)\b|\bctx\.request\.(?:body|query)\b`)
	// A bare `params` is the Next route context; elsewhere it is any helper argument.
	authzBareParams = regexp.MustCompile(`(?:^|[^\w.])params\.\w`)
	authzRoutePath  = regexp.MustCompile(`(?:^|/)route\.[jt]sx?$|(?:^|/)(?:app|pages)/`)
	// A principal or an explicit guard in the same handler counts as an authorization check.
	authzInScope = regexp.MustCompile(`\b(?:req\.user|request\.user|ctx\.user|ctx\.state\.user|session\??\.user|currentUser|authorize(?:d)?|canAccess|hasAccess|assertOwner|belongsTo)\b`)
	// Ownership fields only count when they filter the query itself.
	authzOwnership = regexp.MustCompile(`\b(?:userId|tenantId|ownerId)\b`)

	// Function opener, split in two because JS uses a negative lookahead (RE2 has none).
	fnOpenerSimple = regexp.MustCompile(`=>\s*\{|\bfunction\b[^{]*\{|^\s*\)\s*(?::[^{]*)?\{\s*$`)
	fnOpenerMethod = regexp.MustCompile(`^\s*(?:(?:export|default|public|private|protected|static|async|final|synchronized)\s+)*(?:[\w$.]+(?:<[^()]*>)?(?:\[\])*\s+)?([A-Za-z_$][\w$]*)\s*\(.*\)\s*(?::[^{]*|throws\s+[\w.,\s]+)?\{\s*$`)
	controlWords   = []string{"if", "for", "while", "switch", "catch", "with", "try", "synchronized"}
)

// AuthZ flags object lookups keyed by caller-supplied ids (CWE-639) with no
// ownership filter or principal check in the same handler.
func AuthZ(f diff.File) []finding.Finding {
	return mergeByLine(lookupAuthZ(f), frameworkAuthZ(f))
}

func lookupAuthZ(f diff.File) []finding.Finding {
	callerID := false
	for _, l := range f.Added {
		if !IsCommentLine(l.Text) && isCallerInput(f.Path, l.Text) {
			callerID = true
			break
		}
	}
	if !callerID {
		return nil
	}
	var out []finding.Finding
	for i, l := range f.Lines {
		if l.Kind != diff.Added || IsCommentLine(l.Text) || !authzSink.MatchString(l.Text) {
			continue
		}
		call := sinkCall(f.Lines, i)
		if isPublicLookup(call) || hasAuthzCheck(enclosingHandler(f.Lines, i), call) {
			continue
		}
		out = add(out, finding.Input{
			Title:    "Missing object-level authorization",
			Class:    finding.AuthZ,
			Path:     f.Path,
			Line:     l.Line,
			CWE:      "CWE-639",
			OWASP:    finding.LabelsFor(finding.AuthZ),
			Severity: finding.High,
		})
	}
	return out
}

func hasAuthzCheck(scope []diff.Line, call string) bool {
	if authzOwnership.MatchString(call) {
		return true
	}
	for _, l := range scope {
		if IsCommentLine(l.Text) {
			continue
		}
		if authzInScope.MatchString(stripStrings(l.Text)) || authzAdminGuard.MatchString(l.Text) {
			return true
		}
	}
	return false
}

func isFunctionOpener(text string) bool {
	if fnOpenerSimple.MatchString(text) {
		return true
	}
	m := fnOpenerMethod.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	word := m[1]
	for _, kw := range controlWords {
		if strings.HasPrefix(word, kw) && (len(word) == len(kw) || !isWordByte(word[len(kw)])) {
			return false
		}
	}
	return true
}

// enclosingHandler returns the lines of the innermost function body around
// lines[index], climbing out of if/try/loop blocks and limited to the
// contiguous hunk the sink sits in. Unseen lines are not evidence.
func enclosingHandler(lines []diff.Line, index int) []diff.Line {
	return enclosingBy(lines, index, isFunctionOpener)
}

// enclosingBy is enclosingHandler for any language: isOpener recognises the
// line that opens a function body.
func enclosingBy(lines []diff.Line, index int, isOpener func(string) bool) []diff.Line {
	start, depth, levels, found := index, 0, 0, false
	for i := index - 1; i >= 0; i-- {
		if lines[i].Line != lines[i+1].Line-1 {
			break // hunk boundary
		}
		start = i
		if IsCommentLine(lines[i].Text) {
			continue
		}
		depth -= braceNet(lines[i].Text)
		if depth < -levels {
			levels = -depth // this line opens a block around the sink
			if isOpener(stripStrings(lines[i].Text)) {
				found = true
				break
			}
		}
	}
	// Opener not in the hunk: stop at the first closer of a block we did not see open.
	closeAt := -levels
	if !found {
		closeAt = -levels - 1
	}
	end := index
	depth = braceNet(lines[index].Text)
	for i := index + 1; i < len(lines); i++ {
		if lines[i].Line != lines[i-1].Line+1 {
			break // hunk boundary
		}
		end = i
		if IsCommentLine(lines[i].Text) {
			continue
		}
		depth += braceNet(lines[i].Text)
		if depth <= closeAt {
			break // this line closes the handler
		}
	}
	return lines[start : end+1]
}

// sinkCall returns the text of the sink call, from the sink name until its parentheses balance.
func sinkCall(lines []diff.Line, index int) string {
	var sb strings.Builder
	depth := 0
	for i := index; i < len(lines); i++ {
		if i > index && lines[i].Line != lines[i-1].Line+1 {
			break // hunk boundary
		}
		source := lines[i].Text
		if i == index {
			if loc := authzSink.FindStringIndex(source); loc != nil {
				source = source[loc[0]:]
			}
		}
		code := stripStrings(source)
		for j := 0; j < len(code); j++ {
			ch := code[j]
			sb.WriteByte(ch)
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				if depth == 0 {
					return sb.String()
				}
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
