package rules

import "regexp"

var (
	// Role guards must name an admin-like role: the string is read from the raw
	// line because stripStrings blanks it elsewhere.
	authzAdminGuard = regexp.MustCompile(`(?i)\b(?:requireAdmin|ensureAdmin|isAdmin|adminOnly|adminGuard|adminRequired)\b|\b(?:requireRole|hasRole|checkRole|authorizeRoles?|roles?)\s*\(\s*\[?\s*["'](?:admin|superadmin|staff)\b|@Roles\s*\(\s*["'](?:admin|superadmin|staff)\b`)
	// A lookup that filters on a public-visibility flag or on a slug serves
	// content that is public by design, so there is no owner to check.
	authzPublicFlag = regexp.MustCompile(`(?i)\b(?:is_?)?(?:published|public|visible)\s*[:=]\s*true\b|\bfindBy\w*And(?:Is)?(?:Published|Public|Visible)True\b|(?:->|::)where\(\s*['"](?:is_)?(?:published|public|visible)['"]\s*,\s*(?:true|1)\s*\)`)
	authzSlugOnly   = regexp.MustCompile(`^[^{]*\{\s*(?:where\s*:\s*\{\s*)?slug\s*:[^,{}]*\}`)
)

// isPublicLookup reports whether the sink call looks up public content.
func isPublicLookup(call string) bool {
	return authzPublicFlag.MatchString(call) || authzSlugOnly.MatchString(call)
}

// isCallerInput reports whether a line reads caller-controlled request input.
func isCallerInput(path, text string) bool {
	return authzCallerID.MatchString(text) || authzRoutePath.MatchString(path) && authzBareParams.MatchString(text)
}
