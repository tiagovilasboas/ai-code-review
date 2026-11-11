package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func cfMed(line int) []cfWant {
	return []cfWant{{line, finding.CloudflareConfig, finding.Medium, finding.Review}}
}

const zoneHdr = ` resource "cloudflare_zone_settings_override" "z" {`

func TestCloudflareTerraformSettings(t *testing.T) {
	runCF(t, []cfCase{
		{"ssl off", "zone.tf", []string{zoneHdr, `+    ssl = "off"`}, []cfWant{{2, finding.CloudflareConfig, finding.High, finding.Block}}},
		{"ssl flexible", "zone.tf", []string{zoneHdr, `+    ssl = "flexible"`}, cfMed(2)},
		{"always_use_https off", "zone.tf", []string{zoneHdr, `+    always_use_https = "off"`}, cfMed(2)},
		{"min tls 1.0", "zone.tf", []string{zoneHdr, `+    min_tls_version = "1.0"`}, cfMed(2)},
		{"min tls 1.1", "zone.tf", []string{zoneHdr, `+    min_tls_version = "1.1"`}, cfMed(2)},
		{"security level", "zone.tf", []string{zoneHdr, `+    security_level = "essentially_off"`}, cfMed(2)},
		{"browser check", "zone.tf", []string{zoneHdr, `+    browser_check = "off"`}, cfMed(2)},
		{"bot fight off", "bots.tf", []string{` resource "cloudflare_bot_management" "b" {`, `+  fight_mode = false`}, cfMed(2)},
		{"v5 setting_id form", "zone.tf", []string{` resource "cloudflare_zone_setting" "ssl" {`, `   setting_id = "ssl"`, `+  value      = "flexible"`}, cfMed(3)},
		{"v5 min tls", "zone.tf", []string{` resource "cloudflare_zone_setting" "tls" {`, `   setting_id = "min_tls_version"`, `+  value      = "1.1"`}, cfMed(3)},
		// near misses
		{"ssl full strict", "zone.tf", []string{zoneHdr, `+    ssl = "strict"`}, nil},
		{"min tls 1.2", "zone.tf", []string{zoneHdr, `+    min_tls_version = "1.2"`}, nil},
		{"always_use_https on", "zone.tf", []string{zoneHdr, `+    always_use_https = "on"`}, nil},
		{"security level medium", "zone.tf", []string{zoneHdr, `+    security_level = "medium"`}, nil},
		{"fight mode on", "bots.tf", []string{` resource "cloudflare_bot_management" "b" {`, `+  fight_mode = true`}, nil},
		{"ssl in another provider", "db.tf", []string{` resource "aws_db_instance" "d" {`, `+  ssl = "off"`}, nil},
		{"context only", "zone.tf", []string{zoneHdr, `   ssl = "off"`, `+    tls_1_3 = "on"`}, nil},
		{"comment", "zone.tf", []string{zoneHdr, `+    # ssl = "off"`}, nil},
		{"setting id value ok", "zone.tf", []string{` resource "cloudflare_zone_setting" "ssl" {`, `   setting_id = "ssl"`, `+  value      = "strict"`}, nil},
	})
}

const (
	wafHdr  = ` resource "cloudflare_ruleset" "waf" {`
	wafKind = `   phase = "http_request_firewall_managed"`
	rlHdr   = ` resource "cloudflare_ruleset" "rl" {`
	rlPhase = `   phase = "http_ratelimit"`
)

func TestCloudflareTerraformRulesets(t *testing.T) {
	runCF(t, []cfCase{
		{"managed rule disabled", "waf.tf", []string{wafHdr, wafKind, `   rules {`, `     action = "execute"`, `+    enabled = false`}, cfMed(5)},
		{"managed skip whole ruleset", "waf.tf", []string{wafHdr, wafKind, `   rules {`, `+    action = "skip"`, `     action_parameters {`, `       ruleset = "current"`}, cfMed(4)},
		{"rate limit rule disabled", "rl.tf", []string{rlHdr, rlPhase, `   rules {`, `+    enabled = false`}, []cfWant{{4, finding.CloudflareConfig, finding.Medium, finding.Review}}},
		{"rate limit phase deleted", "rl.tf", []string{rlHdr, `-  phase = "http_ratelimit"`, `   kind  = "zone"`}, cfMed(2)},
		// near misses
		{"skip specific rules only", "waf.tf", []string{wafHdr, wafKind, `   rules {`, `+    action = "skip"`, `     action_parameters {`, `       rules = { "abc" = "def" }`}, nil},
		{"override tuning", "waf.tf", []string{wafHdr, wafKind, `   rules {`, `     action_parameters {`, `       overrides {`, `+        enabled = false`}, nil},
		{"custom phase disabled", "fw.tf", []string{` resource "cloudflare_ruleset" "fw" {`, `   phase = "http_request_firewall_custom"`, `+  enabled = false`}, nil},
		{"non-ruleset resource", "x.tf", []string{` resource "cloudflare_record" "r" {`, `   phase = "http_request_firewall_managed"`, `+  enabled = false`}, nil},
		{"phase not visible", "waf.tf", []string{`+    enabled = false`}, nil},
		{"enabled true", "waf.tf", []string{wafHdr, wafKind, `+  enabled = true`}, nil},
		{"rate limit moved in file", "rl.tf", []string{rlHdr, `+  phase = "http_ratelimit"`, `-  phase = "http_ratelimit"`}, nil},
		{"other phase deleted", "fw.tf", []string{` resource "cloudflare_ruleset" "fw" {`, `-  phase = "http_request_firewall_custom"`, `   kind = "zone"`}, nil},
	})
}
