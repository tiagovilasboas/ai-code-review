package rules

import (
	"path"
	"regexp"
	"strings"
)

var (
	djangoTemplateExt = map[string]bool{".html": true, ".htm": true, ".jinja": true, ".jinja2": true, ".j2": true, ".djhtml": true}
	djangoVariable    = regexp.MustCompile(`\{\{(.*?)\}\}`)
	djangoAutoescape  = regexp.MustCompile(`\{%-?\s*autoescape\s+(?:off|false)\s*-?%\}`)
	djangoMarkSafe    = regexp.MustCompile(`\b(?:mark_safe|Markup)\s*\(`)
	djangoStaticHead  = regexp.MustCompile(`^(?:STATIC_URL|MEDIA_URL|settings\.\w+|static\b)`)
	djangoEscapeCall  = regexp.MustCompile(`\b(?:escape|conditional_escape|format_html|format_html_join|strip_tags|json_script|linebreaks|urlize)\s*\(`)
)

// djangoSafeHit reports an added line that turns off auto-escaping for
// content that is not a constant: |safe, {% autoescape off %} or mark_safe.
func djangoSafeHit(p, text string) bool {
	switch {
	case djangoTemplateExt[strings.ToLower(path.Ext(p))]:
		return djangoAutoescape.MatchString(text) || templateSafeFilter(text)
	case langOf(p) == langPython:
		return markSafeOfDynamic(text)
	}
	return false
}

// templateSafeFilter reports a {{ value|safe }} whose value is neither a
// literal, a static path, nor escaped before the safe filter.
func templateSafeFilter(text string) bool {
	for _, m := range djangoVariable.FindAllStringSubmatch(text, -1) {
		parts := strings.Split(m[1], "|")
		if len(parts) < 2 {
			continue
		}
		head, filters := strings.TrimSpace(parts[0]), parts[1:]
		if !hasFilter(filters, "safe", "safeseq") {
			continue
		}
		if IsStringLiteral(head) || djangoStaticHead.MatchString(head) || escapedBeforeSafe(filters) {
			continue
		}
		return true
	}
	return false
}

func hasFilter(filters []string, names ...string) bool {
	for _, f := range filters {
		name := strings.TrimSpace(strings.SplitN(f, ":", 2)[0])
		for _, n := range names {
			if name == n {
				return true
			}
		}
	}
	return false
}

// escapedBeforeSafe: an escape filter runs before the safe filter.
func escapedBeforeSafe(filters []string) bool {
	for _, f := range filters {
		name := strings.TrimSpace(strings.SplitN(f, ":", 2)[0])
		if name == "safe" || name == "safeseq" {
			return false
		}
		if name == "escape" || name == "force_escape" || name == "escapejs" || name == "urlencode" {
			return true
		}
	}
	return false
}

// markSafeOfDynamic reports mark_safe(x) where x is not a constant string and
// is not itself escaped or built by format_html.
func markSafeOfDynamic(text string) bool {
	loc := djangoMarkSafe.FindStringIndex(text)
	if loc == nil {
		return false
	}
	args := splitCallArgs(text[loc[1]:])
	if len(args) == 0 {
		return false
	}
	arg := ssrfStripPrefix(args[0])
	if isPureLiteral(arg) && !strings.HasPrefix(strings.TrimSpace(args[0]), "f") {
		return false
	}
	return !coveredByCall(arg, djangoEscapeCall)
}
