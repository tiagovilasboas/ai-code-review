package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	wranglerFile = regexp.MustCompile(`^wrangler[\w.-]*\.(?:toml|jsonc?)$`)
	tomlTable    = regexp.MustCompile(`^\s*\[\[?\s*([\w.\-"]+)\s*\]\]?\s*(?:#.*)?$`)
	tomlKV       = regexp.MustCompile(`^\s*["']?([\w.\-]+)["']?\s*=\s*(.+?)\s*$`)
	jsonKV       = regexp.MustCompile(`^\s*"([\w.\-]+)"\s*:\s*(.+?)\s*,?\s*$`)
	jsonVarsOpen = regexp.MustCompile(`"vars"\s*:\s*\{`)
	// Names that hold a credential. SITE_KEY and PUBLIC_KEY are public by design.
	wranglerSecretName = regexp.MustCompile(`(?i)secret|token|passw(?:or)?d|credential|private[_-]?key|api[_-]?key|access[_-]?key|signing[_-]?key|auth[_-]?key|webhook[_-]?key`)
	wranglerPublicName = regexp.MustCompile(`(?i)site[_-]?key|public[_-]?key`)
	wranglerLiveShape  = regexp.MustCompile(`^(?:sk_live_|rk_live_|ghp_|github_pat_|xox[baprs]-|AKIA[0-9A-Z]{12,})`)
	publicURLSetting   = regexp.MustCompile(`(?i)(?:workers_dev|preview_urls)["']?\s*[=:]\s*true\b`)
)

// wranglerFindings checks wrangler.toml / wrangler.json(c): a credential under
// vars (it ships in the bundle and the dashboard) and a public workers.dev or
// preview URL added.
func wranglerFindings(f diff.File) []finding.Finding {
	if !wranglerFile.MatchString(baseLower(f.Path)) {
		return nil
	}
	isTOML := strings.HasSuffix(baseLower(f.Path), ".toml")
	state := varsUnknown
	depth := 0
	var out []finding.Finding
	for _, l := range f.Lines {
		if isTOML {
			state = tomlSection(state, l.Text)
		} else {
			state, depth = jsonSection(state, depth, l.Text)
		}
		if l.Kind != diff.Added || IsCommentLine(l.Text) {
			continue
		}
		if publicURLSetting.MatchString(l.Text) {
			out = add(out, cfConfig(f, l.Line, "workers.dev or preview URL enabled", "CWE-288", finding.Medium, "A01:2021"))
			continue
		}
		if state == varsOut {
			continue
		}
		if name, val, ok := wranglerPair(isTOML, l.Text); ok && wranglerHoldsSecret(name, val) {
			out = add(out, cfSecret(f, l.Line, "Credential in wrangler vars"))
		}
	}
	return out
}

type varsState int

const (
	varsUnknown varsState = iota // the hunk does not show the enclosing table
	varsIn
	varsOut
)

func tomlSection(cur varsState, text string) varsState {
	m := tomlTable.FindStringSubmatch(text)
	if m == nil {
		return cur
	}
	t := strings.Trim(m[1], `"`)
	if t == "vars" || strings.HasSuffix(t, ".vars") {
		return varsIn
	}
	return varsOut
}

// jsonSection tracks whether a line sits inside the "vars" object by brace depth.
func jsonSection(cur varsState, depth int, text string) (varsState, int) {
	if jsonVarsOpen.MatchString(text) {
		if depth = strings.Count(text, "{") - strings.Count(text, "}"); depth <= 0 {
			return varsOut, 0
		}
		return varsIn, depth
	}
	if cur != varsIn {
		return cur, depth
	}
	depth += strings.Count(text, "{") - strings.Count(text, "}")
	if depth <= 0 {
		return varsOut, 0
	}
	return varsIn, depth
}

// wranglerPair reads `name = value` (TOML, also inline tables) or `"name": value` (JSON).
func wranglerPair(isTOML bool, text string) (name, val string, ok bool) {
	re := jsonKV
	if isTOML {
		re = tomlKV
	}
	m := re.FindStringSubmatch(text)
	if m == nil {
		return "", "", false
	}
	val = m[2]
	if isTOML {
		val = stripLineComment(val)
	}
	if len(val) < 2 || (val[0] != '"' && val[0] != '\'') {
		return "", "", false
	}
	return m[1], strings.Trim(val, `"',`), true
}

func wranglerHoldsSecret(name, val string) bool {
	if wranglerLiveShape.MatchString(val) {
		return true
	}
	if wranglerPublicName.MatchString(name) || !wranglerSecretName.MatchString(name) {
		return false
	}
	return isLiteralSecretValue(val) && !isNameLiteral(val)
}
