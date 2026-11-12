package rules

import "regexp"

var pythonWriter = htmlWriter{
	// A returned view value, or the body handed to a response class.
	sink:   regexp.MustCompile(`(?:\breturn|\b(?:HttpResponse|HTMLResponse|make_response|render_template_string|Response)\()\s*(?:content\s*=\s*)?(.*)$`),
	escape: regexp.MustCompile(`(?:html\.escape|markupsafe\.escape|escape|conditional_escape|format_html|bleach\.clean|strip_tags|quote|urlencode|int|float)\s*\(`),
	skip:   regexp.MustCompile(`\bjsonify\s*\(|JSONResponse|json\.dumps|\breturn\s+\{|(?:mimetype|content_type|media_type)\s*=\s*["'](?:text/plain|application/json|[^"']*javascript)`),
}
