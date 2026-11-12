package rules

import (
	"path"
	"regexp"
	"strings"
)

var (
	phpWriter = htmlWriter{
		sink:       regexp.MustCompile(`(?:\b(?:echo|print|printf)\b|<\?=)\s*\(?(.*)$`),
		escape:     regexp.MustCompile(`(?:htmlspecialchars|htmlentities|esc_html|esc_attr|esc_url|e|strip_tags|intval|floatval|absint|json_encode|urlencode|rawurlencode|Purifier::clean|count)\s*\(`),
		notHTML:    regexp.MustCompile(`(?i)header\(\s*['"]Content-Type:\s*(?:application/json|text/plain)`),
		php:        true,
		bare:       true,
		formatArgs: true,
	}
	bladeRaw      = regexp.MustCompile(`\{!!\s*(.*?)\s*!!\}`)
	bladeEscaped  = regexp.MustCompile(`\b(?:e|htmlspecialchars|htmlentities|esc_html|strip_tags|Purifier::clean)\s*\(`)
	bladeConstant = regexp.MustCompile(`^(?:__|trans|trans_choice|route|url|asset|mix|vite|csrf_field|csrf_token|method_field|json_encode|Js::from|\$slot\b|\$attributes\b|\$errors\b|\$\w+->(?:links|render)\s*\()`)
	twigAutoesc   = regexp.MustCompile(`\{%-?\s*autoescape\s+(?:false|['"]false['"])`)
	twigEscapers  = map[string]bool{"escape": true, "e": true, "striptags": true, "url_encode": true, "json_encode": true}
)

// phpTemplateHit reports Blade {!! !!} and Twig |raw over a non-constant value.
func phpTemplateHit(p, text string) bool {
	switch {
	case strings.HasSuffix(strings.ToLower(p), ".blade.php"):
		for _, m := range bladeRaw.FindAllStringSubmatch(text, -1) {
			if !IsStringLiteral(m[1]) && !bladeConstant.MatchString(m[1]) && !bladeEscaped.MatchString(m[1]) {
				return true
			}
		}
	case strings.ToLower(path.Ext(p)) == ".twig":
		return twigAutoesc.MatchString(text) || twigRawFilter(text)
	}
	return false
}

func twigRawFilter(text string) bool {
	for _, m := range djangoVariable.FindAllStringSubmatch(text, -1) {
		parts := strings.Split(m[1], "|")
		head := strings.TrimSpace(parts[0])
		if IsStringLiteral(head) || djangoStaticHead.MatchString(head) {
			continue
		}
		for _, f := range parts[1:] {
			name := strings.TrimSpace(strings.SplitN(f, "(", 2)[0])
			if twigEscapers[name] {
				break
			}
			if name == "raw" {
				return true
			}
		}
	}
	return false
}
