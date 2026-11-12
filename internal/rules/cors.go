package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	corsStar    = regexp.MustCompile(`\bcors\s*\(\s*\{[^}]*origin\s*:\s*["']\*["']`)
	corsOptStar = regexp.MustCompile(`\borigin\s*:\s*["']\*["']`)
	corsMention = regexp.MustCompile(`(?i)cors`)
	// Allow-Origin header (set by any framework or server config) whose value is "*".
	corsHeaderStar = regexp.MustCompile(`(?i)access-control-allow-origin[^*\n]{0,30}?["'\s=:,]\*(?:["'\s;,)}]|$)`)
	corsHeaderName = regexp.MustCompile(`(?i)access-control-allow-origin`)
	// The request's Origin echoed back, as a header value or as `origin: true`.
	corsReflectHeader = regexp.MustCompile(`(?i)headers?\.origin|headers\[\s*["']origin["']\s*\]|\.(?:get|header|getHeader)\(\s*["']origin["']\s*[,)]` + corsLangOrigin)
	corsReflectOpt    = regexp.MustCompile(`\borigin\s*:\s*true\b`)
	corsCredentials   = regexp.MustCompile(`(?i)\bcredentials\s*:\s*true\b|access-control-allow-credentials["'\s,:=]+["']?true|\ballowCredentials\s*(?:=\s*"?|\(\s*)true|` + corsLangCredentials)
	corsMethods       = regexp.MustCompile(`(?i)(?:\b|_)methods?\b|\ballowMethod\b`)
	corsVerb          = regexp.MustCompile(`(?i)\b(GET|HEAD|OPTIONS|POST|PUT|PATCH|DELETE)\b`)
	corsStaticAsset   = regexp.MustCompile(`express\.static|serveStatic|["'/](?:fonts|assets|static)\b`)
)

// corsFindings flags permissive CORS (CWE-942): a wildcard origin, or the
// request origin reflected while credentials are allowed. A wildcard without
// credentials that only serves read-only methods or static assets is public
// by design and stays silent.
func corsFindings(f diff.File) []finding.Finding {
	code := nonComment(f.Added)
	mentionsCORS, credsAnywhere := false, false
	for _, l := range code {
		mentionsCORS = mentionsCORS || corsMention.MatchString(l.Text)
		credsAnywhere = credsAnywhere || corsCredentials.MatchString(l.Text)
	}
	publicReadOnly := !credsAnywhere && (readOnlyMethods(code) || anyLine(code, corsStaticAsset))
	vars := originVars(code)
	var out []finding.Finding
	for _, l := range code {
		wild := corsStar.MatchString(l.Text) || corsHeaderStar.MatchString(l.Text) ||
			corsSpringStar.MatchString(l.Text) || corsLangStar.MatchString(l.Text) || corsFlaskDefaultOrigins(l.Text) || (mentionsCORS && corsOptStar.MatchString(l.Text))
		reflect := corsReflectOpt.MatchString(l.Text) && mentionsCORS && credentialsNear(code, l.Line) ||
			echoesRequestOrigin(l, nonComment(f.Lines), vars) && (credentialsNear(code, l.Line) || !publicReadOnly)
		if !(wild && !publicReadOnly) && !reflect {
			continue
		}
		out = add(out, finding.Input{
			Title:    "Overly permissive CORS origin",
			Class:    finding.CORS,
			Path:     f.Path,
			Line:     l.Line,
			CWE:      "CWE-942",
			OWASP:    finding.LabelsFor(finding.CORS),
			Severity: finding.Medium,
		})
	}
	return out
}

func nonComment(lines []diff.Line) []diff.Line {
	var out []diff.Line
	for _, l := range lines {
		if !IsCommentLine(l.Text) {
			out = append(out, l)
		}
	}
	return out
}

func anyLine(lines []diff.Line, re *regexp.Regexp) bool {
	for _, l := range lines {
		if re.MatchString(l.Text) {
			return true
		}
	}
	return false
}

// credentialsNear reports an allow-credentials setting within four lines.
func credentialsNear(code []diff.Line, line int) bool {
	for _, l := range code {
		if l.Line >= line-4 && l.Line <= line+4 && corsCredentials.MatchString(l.Text) {
			return true
		}
	}
	return false
}

// readOnlyMethods reports a methods list that names only safe verbs.
func readOnlyMethods(code []diff.Line) bool {
	seen := false
	for _, l := range code {
		if !corsMethods.MatchString(l.Text) {
			continue
		}
		for _, m := range corsVerb.FindAllStringSubmatch(l.Text, -1) {
			switch strings.ToUpper(m[1]) {
			case "GET", "HEAD", "OPTIONS":
				seen = true
			default:
				return false
			}
		}
	}
	return seen
}
