package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	jaxrsRoute = regexp.MustCompile(`^\s*@(?:GET|POST|PUT|PATCH|DELETE)\s*$`)
	// @PathParam("id") Long id; group 1 is the variable.
	jaxrsPathParam = regexp.MustCompile(`@PathParam\b(?:\s*\([^)]*\))?\s+(?:final\s+)?[\w.<>\[\]]+\s+(\w+)`)
	jaxrsSink      = regexp.MustCompile(`\.(?:findById|getById|getReference|find|load|get\w*ById|find\w*By\w*|delete\w*|remove\w*|update\w*|merge|save)\(`)
	// Role annotations, the injected SecurityContext, an ownership field or an explicit 403.
	jaxrsPrincipal = regexp.MustCompile(`@(?:RolesAllowed|PreAuthorize|Secured|DenyAll|Authenticated)\b|\b(?:SecurityContext|Principal|getUserPrincipal|isUserInRole|ForbiddenException|NotAuthorizedException)\b|\bResponse\.Status\.FORBIDDEN\b|\b(?:userId|ownerId|tenantId)\b|(?i:currentUser|isOwner)|\bfind\w*And(?:User|Owner|Tenant)\w*\(`)
)

// jaxrsUnchecked returns the first id-keyed lookup or mutation of every
// JAX-RS resource method that takes a @PathParam id and shows no role
// annotation, SecurityContext use or ownership filter.
func jaxrsUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for i, l := range lines {
		if l.Kind != diff.Added || !jaxrsRoute.MatchString(l.Text) {
			continue
		}
		method := braceBlock(lines, i, 12)
		if method == nil {
			continue
		}
		sink, ok := firstSink(method, jaxrsSink, jaxrsIDs(method))
		if !ok {
			continue
		}
		scope := append(append(decoratorsAbove(lines, i), springClassAnnotations(lines, i)...), jaxrsBody(method)...)
		if !hasGuard(scope, jaxrsPrincipal) {
			out = append(out, sink)
		}
	}
	return out
}

// jaxrsIDs matches the @PathParam variables that are not a slug.
func jaxrsIDs(method []diff.Line) []*regexp.Regexp {
	var ids []*regexp.Regexp
	for _, l := range method {
		for _, m := range jaxrsPathParam.FindAllStringSubmatch(l.Text, -1) {
			if m[1] != "slug" {
				ids = append(ids, wordRegexp(m[1]))
			}
		}
	}
	return ids
}

// jaxrsBody removes the @PathParam parameters from the method lines: a
// `Long userId` path param is the caller's input, not a principal. The
// injected SecurityContext parameter stays in the scope on purpose.
func jaxrsBody(method []diff.Line) []diff.Line {
	rest := make([]diff.Line, len(method))
	for i, l := range method {
		l.Text = jaxrsPathParam.ReplaceAllString(l.Text, "")
		rest[i] = l
	}
	return rest
}
