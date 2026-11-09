package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	nestRoute = regexp.MustCompile(`^\s*@(?:Get|Post|Put|Patch|Delete|All)\s*\(`)
	nestClass = regexp.MustCompile(`\bclass\s+\w+`)
	// @Param('id') id, @Query('id') id; group 1 is the name, group 2 the variable.
	nestParam = regexp.MustCompile(`@(?:Param|Query)\(\s*(?:['"](\w+)['"])?\s*\)\s*(\w+)`)
	nestBody  = regexp.MustCompile(`@Body\(\s*(?:['"]\w+['"])?\s*\)\s*(\w+)`)
	nestSink  = regexp.MustCompile(`\.(?:update|find|get|remove|delete|softDelete|restore|patch|save|replace|load)\w*\(`)
	// The request user, via @Req()/@CurrentUser() style decorators or fields.
	nestPrincipal = regexp.MustCompile(`@(?:Req|Request|CurrentUser|User|GetUser|AuthUser|ActiveUser|Session|AuthenticatedUser)\s*\(|\b(?:req|request)\.user\b|(?i:currentUser)|\buserId\b|\bownerId\b|\btenantId\b|\buser\.id\b`)
	// Guards and decorators that check ownership or permission, or open the route on purpose.
	nestGuard = regexp.MustCompile(`@UseGuards\s*\([^)]*(?:Owner|Ownership|Policy|Policies|Ability|Abilities|Permission|Casl|Acl|Admin)\w*|@(?:CheckPolicies|Policies|Can|Authorize|Permissions|RequirePermissions|Public)\b`)
)

// nestUnchecked returns the first id-keyed lookup line of every NestJS route
// method that has no user or ownership guard in scope.
func nestUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for i, l := range lines {
		if l.Kind != diff.Added || !nestRoute.MatchString(l.Text) {
			continue
		}
		method := braceBlock(lines, i, 12)
		if method == nil {
			continue
		}
		ids := nestCallerIDs(method)
		sink, ok := firstSink(method, nestSink, ids)
		if !ok {
			continue
		}
		scope := append(append(decoratorsAbove(lines, i), classDecorators(lines, i)...), method...)
		if !hasGuard(scope, nestPrincipal) && !anyLine(scope, nestGuard) {
			out = append(out, sink)
		}
	}
	return out
}

// nestCallerIDs matches route/query parameter variables (never a slug) and
// the id field of a request body DTO.
func nestCallerIDs(method []diff.Line) []*regexp.Regexp {
	var ids []*regexp.Regexp
	for _, l := range method {
		for _, m := range nestParam.FindAllStringSubmatch(l.Text, -1) {
			if m[1] != "slug" {
				ids = append(ids, wordRegexp(m[2]))
			}
		}
		for _, m := range nestBody.FindAllStringSubmatch(l.Text, -1) {
			ids = append(ids, regexp.MustCompile(`\b`+regexp.QuoteMeta(m[1])+`\.\w*[iI]d\b`))
		}
	}
	return ids
}

// decoratorsAbove returns the single-line decorators stacked directly above lines[index].
func decoratorsAbove(lines []diff.Line, index int) []diff.Line {
	start := index
	for i := index - 1; i >= 0 && lines[i].Line == lines[i+1].Line-1; i-- {
		if !strings.HasPrefix(strings.TrimSpace(lines[i].Text), "@") {
			break
		}
		start = i
	}
	return lines[start:index]
}

// classDecorators returns the decorators of the nearest class above lines[index].
func classDecorators(lines []diff.Line, index int) []diff.Line {
	for i := index - 1; i >= 0 && lines[i].Line == lines[i+1].Line-1; i-- {
		if nestClass.MatchString(lines[i].Text) {
			return decoratorsAbove(lines, i)
		}
	}
	return nil
}
