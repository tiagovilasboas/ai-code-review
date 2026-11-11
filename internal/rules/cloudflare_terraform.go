package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// tfSetting is a zone setting that weakens the edge when set to a given value.
type tfSetting struct {
	name, title, cwe string
	bad              map[string]finding.Severity
	extra            string
}

var tfSettings = []tfSetting{
	{"ssl", "Cloudflare SSL mode weakened", "CWE-319", map[string]finding.Severity{"off": finding.High, "flexible": finding.Medium}, "A02:2021"},
	{"always_use_https", "Always Use HTTPS turned off", "CWE-319", map[string]finding.Severity{"off": finding.Medium}, "A02:2021"},
	{"min_tls_version", "Minimum TLS version below 1.2", "CWE-319", map[string]finding.Severity{"1.0": finding.Medium, "1.1": finding.Medium}, "A02:2021"},
	{"security_level", "Security level set to essentially off", "CWE-799", map[string]finding.Severity{"essentially_off": finding.Medium}, ""},
	{"browser_check", "Browser Integrity Check turned off", "CWE-799", map[string]finding.Severity{"off": finding.Medium}, ""},
}

var (
	tfFile     = regexp.MustCompile(`(?i)\.tf$`)
	tfAttr     = regexp.MustCompile(`^\s*([a-z_]+)\s*=\s*"?([\w.]+)"?\s*(?:#.*|//.*)?$`)
	tfFalse    = regexp.MustCompile(`(?i)^\s*(?:fight_mode|bot_fight_mode|super_bot_fight_mode)\s*=\s*false\b`)
	tfSetting5 = regexp.MustCompile(`^\s*setting_id\s*=\s*"([a-z_]+)"`)
	tfCloud    = regexp.MustCompile(`(?i)cloudflare`)
)

// terraformSettingFindings flags a weakened zone setting added in Terraform for
// the cloudflare provider: TLS or HTTPS downgrades, security level, browser
// check, and bot fight mode turned off. `ssl` also exists in other providers,
// so it needs a cloudflare mention in the diff or the path.
func terraformSettingFindings(f diff.File) []finding.Finding {
	if !tfFile.MatchString(f.Path) {
		return nil
	}
	cloud := tfCloud.MatchString(f.Path) || anyLine(f.Lines, tfCloud)
	var out []finding.Finding
	settingID := ""
	for _, l := range f.Lines {
		if strings.HasPrefix(strings.TrimSpace(l.Text), "resource ") {
			settingID = ""
		}
		if m := tfSetting5.FindStringSubmatch(l.Text); m != nil {
			settingID = m[1]
		}
		if l.Kind != diff.Added || IsCommentLine(l.Text) {
			continue
		}
		if tfFalse.MatchString(l.Text) && cloud {
			out = add(out, cfConfig(f, l.Line, "Bot Fight Mode turned off", "CWE-799", finding.Medium))
			continue
		}
		m := tfAttr.FindStringSubmatch(l.Text)
		if m == nil {
			continue
		}
		name, val := m[1], strings.ToLower(m[2])
		if name == "value" && settingID != "" {
			name = settingID // cloudflare_zone_setting: setting_id = "ssl", value = "off"
		}
		if in, ok := weakSetting(f, l.Line, name, val, cloud); ok {
			out = add(out, in)
		}
	}
	return out
}

func weakSetting(f diff.File, line int, name, val string, cloud bool) (finding.Input, bool) {
	for _, s := range tfSettings {
		sev, bad := s.bad[val]
		if s.name != name || !bad || (name == "ssl" && !cloud) {
			continue
		}
		extra := []string{}
		if s.extra != "" {
			extra = append(extra, s.extra)
		}
		return cfConfig(f, line, s.title, s.cwe, sev, extra...), true
	}
	return finding.Input{}, false
}
