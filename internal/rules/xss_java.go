package rules

import (
	"path"
	"regexp"
	"strings"
)

var (
	javaWriter = htmlWriter{
		// Servlet writers, a Vert.x response().end(...), or a Spring handler returning the string.
		sink: regexp.MustCompile(`(?:\b(?:out|writer|pw|\w+\.getWriter\(\))\.(?:print|println|write|append|printf)\s*\(` +
			`|\bresponse\(\)(?:\.\w+\([^()]*\))*\.(?:end|write|send)\s*\(|\breturn\s+)(.*)$`),
		escape:     regexp.MustCompile(`(?:\w+\.)*(?:escapeHtml\d?|escapeXml\w*|htmlEscape|htmlEscaper\(\)\.escape|forHtml\w*|encodeForHTML|encodeForHtml|sanitize|clean)\s*\(|ESAPI\.encoder\(\)\.encodeForHTML\s*\(`),
		format:     regexp.MustCompile(`\bString\.format\(\s*(.*)$`),
		skip:       regexp.MustCompile(`\.toJson\(|\bwriteValueAsString\(`),
		notHTML:    regexp.MustCompile(`(?i)(?:setContentType\(|putHeader\(\s*"content-type"\s*,|CONTENT_TYPE\s*,)\s*"(?:application/json|text/plain)`),
		formatArgs: true,
	}
	jspOutput   = regexp.MustCompile(`<%=\s*(.*?)\s*%>`)
	jspParamEL  = regexp.MustCompile(`\$\{\s*(?:param|paramValues|header|cookie)\b`)
	jspTag      = regexp.MustCompile(`<c:`)
	jspRawOut   = regexp.MustCompile(`<c:out\b[^>]*\bvalue\s*=\s*"\$\{`)
	jspNoEscape = regexp.MustCompile(`escapeXml\s*=\s*"false"`)
	javaEscaped = regexp.MustCompile(`escapeXml|escapeHtml|htmlEscape|Encode\.for|ESAPI|sanitize`)
	thUtext     = regexp.MustCompile(`\bth:utext\s*=\s*"([^"]*)"|\[\(\s*(.*?)\s*\)\]`)
)

// javaTemplateHit reports JSP output and Thymeleaf unescaped text that carry
// a non-constant value.
func javaTemplateHit(p, text string) bool {
	switch strings.ToLower(path.Ext(p)) {
	case ".jsp", ".jspx", ".tag":
		return jspHit(text)
	case ".html", ".xhtml":
		for _, m := range thUtext.FindAllStringSubmatch(text, -1) {
			v := m[1] + m[2]
			if strings.Contains(v, "${") && !javaEscaped.MatchString(v) {
				return true
			}
		}
	}
	return false
}

func jspHit(text string) bool {
	if javaEscaped.MatchString(text) && !jspNoEscape.MatchString(text) {
		return false
	}
	if m := jspOutput.FindStringSubmatch(text); m != nil && requestInput.MatchString(m[1]) {
		return true
	}
	if jspRawOut.MatchString(text) {
		return jspNoEscape.MatchString(text)
	}
	return !jspTag.MatchString(text) && jspParamEL.MatchString(text)
}
