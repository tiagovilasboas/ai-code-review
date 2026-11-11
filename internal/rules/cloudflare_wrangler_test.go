package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var secretBlock = cfWant{2, finding.Secret, finding.Critical, finding.Block}

func TestCloudflareWranglerVars(t *testing.T) {
	runCF(t, []cfCase{
		{"toml secret in vars", "wrangler.toml", []string{"+[vars]", `+STRIPE_SECRET_KEY = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`}, []cfWant{secretBlock}},
		{"toml lower-case name", "wrangler.toml", []string{"+[vars]", `+payment_api_key = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`}, []cfWant{secretBlock}},
		{"toml env vars table", "wrangler.toml", []string{"+[env.production.vars]", `+WEBHOOK_SECRET = "whsec_Zq81xPmL3vT7nB5c"`}, []cfWant{secretBlock}},
		{"toml known prefix any name", "wrangler.toml", []string{"+[vars]", `+STRIPE = "sk_live_Zq81xPmL3vT7nB5cD9fH"`}, []cfWant{secretBlock}},
		{"toml header outside hunk", "wrangler.toml", []string{`+CF_API_TOKEN = "Zq81xPmL3vT7nB5cD9fHk3Jx9Qm2Lp7Vw5Rt8Yu1"`}, []cfWant{{1, finding.Secret, finding.Critical, finding.Block}}},
		{"jsonc secret in vars", "wrangler.jsonc", []string{`+  "vars": {`, `+    "AUTH_TOKEN": "k3Jx9Qm2Lp7Vw5Rt8Yu1"`, `+  }`}, []cfWant{secretBlock}},
		{"json vars with context opener", "wrangler.json", []string{`   "vars": {`, `+    "DB_PASSWORD": "Tr0ub4dor83xQ"`}, []cfWant{secretBlock}},
		// near misses
		{"placeholder", "wrangler.toml", []string{"+[vars]", `+API_TOKEN = "<your-token>"`}, nil},
		{"env reference", "wrangler.toml", []string{"+[vars]", `+API_TOKEN = "${API_TOKEN}"`}, nil},
		{"empty value", "wrangler.toml", []string{"+[vars]", `+API_TOKEN = ""`}, nil},
		{"turnstile test secret", "wrangler.toml", []string{"+[vars]", `+TURNSTILE_SECRET_KEY = "1x0000000000000000000000000000000AA"`}, nil},
		{"turnstile site key", "wrangler.toml", []string{"+[vars]", `+TURNSTILE_SITE_KEY = "0x4AAAAAAABkMYinukE8nzYS"`}, nil},
		{"non credential var", "wrangler.toml", []string{"+[vars]", `+LOG_LEVEL = "debug"`}, nil},
		{"outside vars table", "wrangler.toml", []string{"+[[kv_namespaces]]", `+binding = "SESSION_TOKEN"`}, nil},
		{"toml comment", "wrangler.toml", []string{"+[vars]", `+# API_TOKEN = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`}, nil},
		{"not a wrangler file", "config/app.toml", []string{"+[vars]", `+payment_api_key = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`}, nil},
		{"context only", "wrangler.toml", []string{" [vars]", ` payment_api_key = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`, "+X = 1"}, nil},
	})
}

func TestCloudflareWorkersDev(t *testing.T) {
	med := func(line int) []cfWant {
		return []cfWant{{line, finding.CloudflareConfig, finding.Medium, finding.Review}}
	}
	runCF(t, []cfCase{
		{"workers_dev true", "wrangler.toml", []string{`+workers_dev = true`}, med(1)},
		{"preview_urls true", "wrangler.toml", []string{`+preview_urls = true`}, med(1)},
		{"json workers_dev", "wrangler.jsonc", []string{`+  "workers_dev": true,`}, med(1)},
		{"workers_dev false", "wrangler.toml", []string{`+workers_dev = false`}, nil},
		{"preview_urls false", "wrangler.toml", []string{`+preview_urls = false`}, nil},
		{"context only", "wrangler.toml", []string{` workers_dev = true`, `+name = "api"`}, nil},
		{"not a wrangler file", "docs/notes.toml", []string{`+workers_dev = true`}, nil},
	})
}
