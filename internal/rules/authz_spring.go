package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	springRoute = regexp.MustCompile(`^\s*@(?:Get|Post|Put|Patch|Delete|Request)Mapping\b`)
	springClass = regexp.MustCompile(`\bclass\s+\w+`)
	// @PathVariable Long id, @RequestParam("id") Long id; group 1 is the
	// annotation, group 2 the variable.
	springIDParam = regexp.MustCompile(`@(PathVariable|RequestParam)\b(?:\s*\([^)]*\))?\s+(?:final\s+)?[\w.<>\[\]]+\s+(\w+)`)
	springSink    = regexp.MustCompile(`\.(?:findById|getById|getReferenceById|getOne|findOne|deleteById|find|load|get\w*ById|find\w*By\w*|delete\w*|update\w*|save)\(`)
	// The authenticated principal, a method/class security annotation, an
	// ownership-scoped repository method or an explicit 403.
	// springWeakRole is a method-security annotation that only asks for a login or
	// a plain user role: any signed-in user passes, so it says nothing about ownership.
	springWeakRole   = regexp.MustCompile(`(?i)@(?:PreAuthorize|PostAuthorize|Secured|RolesAllowed)\b[^\n]*(?:isAuthenticated\(\)|\b(?:hasRole|hasAnyRole|hasAuthority|hasAnyAuthority)\(\s*['"](?:ROLE_)?(?:USER|CUSTOMER|MEMBER|CLIENT)['"]\s*\)|["'](?:ROLE_)?(?:USER|CUSTOMER|MEMBER|CLIENT)["'])`)
	springStrongRole = regexp.MustCompile(`(?i)admin|staff|owner|principal|authentication\.|#|@\w+\.|hasPermission|manager`)
	springPrincipal  = regexp.MustCompile(`@(?:AuthenticationPrincipal|PreAuthorize|PostAuthorize|Secured|RolesAllowed|CurrentSecurityContext)\b|\b(?:Principal|Authentication|SecurityContextHolder|hasPermission|AccessDeniedException|ForbiddenException|ResponseStatusException)\b|(?i:currentUser|getUserId|getOwner|isOwner)|\bfind\w*And(?:User|Owner|Tenant|Org\w*|Customer|Account)\w*\(|\b(?:userId|ownerId|tenantId)\b|\bHttpStatus\.FORBIDDEN\b`)
)

// springUnchecked returns the first id-keyed lookup of every Spring handler
// whose id comes from the URL (or an id request parameter) and that shows no
// principal, ownership filter or security annotation.
func springUnchecked(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for i, l := range lines {
		if !springRoute.MatchString(l.Text) { // the route may be a context line; the sink must be added
			continue
		}
		method := braceBlock(lines, i, 12)
		if method == nil {
			continue
		}
		ids := springCallerIDs(method)
		sink, ok := firstSink(method, springSink, ids)
		if !ok {
			continue
		}
		scope := append(append(decoratorsAbove(lines, i), springClassAnnotations(lines, i)...), springBody(method)...)
		if !hasGuard(withoutWeakRoles(scope), springPrincipal) {
			out = append(out, sink)
		}
	}
	return out
}

// springCallerIDs matches @PathVariable names (never a slug) and
// @RequestParam names that read like an id.
func springCallerIDs(method []diff.Line) []*regexp.Regexp {
	var ids []*regexp.Regexp
	for _, l := range method {
		for _, m := range springIDParam.FindAllStringSubmatch(l.Text, -1) {
			if m[2] != "slug" && (m[1] == "PathVariable" || authzIDName.MatchString(m[2])) {
				ids = append(ids, wordRegexp(m[2]))
			}
		}
	}
	return ids
}

// springBody drops the parameter lines: a `Long userId` path variable is the
// caller's input, not a principal.
func springBody(method []diff.Line) []diff.Line {
	var rest []diff.Line
	for _, l := range method {
		if !springIDParam.MatchString(l.Text) {
			rest = append(rest, l)
		}
	}
	return rest
}

// springClassAnnotations returns the annotations of the nearest class above.
func springClassAnnotations(lines []diff.Line, index int) []diff.Line {
	for i := index - 1; i >= 0 && lines[i].Line == lines[i+1].Line-1; i-- {
		if springClass.MatchString(lines[i].Text) {
			return decoratorsAbove(lines, i)
		}
	}
	return nil
}

// authzIDName matches variable names that read like an object id.
var authzIDName = regexp.MustCompile(`(?i)(?:^id|\w+Id|\w+_id)$`)

// withoutWeakRoles drops annotations that only require a login or a user role.
func withoutWeakRoles(scope []diff.Line) []diff.Line {
	var rest []diff.Line
	for _, l := range scope {
		if !(springWeakRole.MatchString(l.Text) && !springStrongRole.MatchString(l.Text)) {
			rest = append(rest, l)
		}
	}
	return rest
}
