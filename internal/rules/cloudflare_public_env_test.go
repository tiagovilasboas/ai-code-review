package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestCloudflarePublicEnvSecrets(t *testing.T) {
	hit := []cfWant{{1, finding.Secret, finding.Critical, finding.Block}}
	runCF(t, []cfCase{
		{"next public turnstile secret", ".env.production", []string{`+NEXT_PUBLIC_TURNSTILE_SECRET_KEY=0x4AAAAAAABkMYinukE8nzYSZq81xPmL3v`}, hit},
		{"vite api token", ".env", []string{`+VITE_CLOUDFLARE_API_TOKEN=Zq81xPmL3vT7nB5cD9fHk3Jx9Qm2Lp7Vw5Rt8Yu1`}, hit},
		{"public prefix cf token", ".env.local", []string{`+PUBLIC_CF_API_TOKEN="Zq81xPmL3vT7nB5cD9fHk3Jx9Qm2Lp7Vw5Rt8Yu1"`}, hit},
		{"tsx literal", "src/components/Form.tsx", []string{`+const turnstileSecret = "0x4AAAAAAABkMYinukE8nzYSZq81xPmL3v";`}, hit},
		// near misses
		{"site key is public", ".env", []string{`+NEXT_PUBLIC_TURNSTILE_SITE_KEY=0x4AAAAAAABkMYinukE8nzYS`}, nil},
		{"test secret", ".env", []string{`+NEXT_PUBLIC_TURNSTILE_SECRET_KEY=1x0000000000000000000000000000000AA`}, nil},
		{"placeholder", ".env.example", []string{`+VITE_TURNSTILE_SECRET=your-secret-here`}, nil},
		{"empty", ".env", []string{`+VITE_TURNSTILE_SECRET=`}, nil},
		{"reference", "src/env.ts", []string{`+VITE_TURNSTILE_SECRET: process.env.TURNSTILE_SECRET,`}, nil},
		{"private server var", ".env", []string{`+TURNSTILE_SECRET_KEY=0x4AAAAAAABkMYinukE8nzYSZq81xPmL3v`}, nil},
		{"server file literal not client", "src/worker/verify.ts", []string{`+const turnstileSecret = "0x4AAAAAAABkMYinukE8nzYSZq81xPmL3v";`}, nil},
		{"comment", ".env", []string{`+# NEXT_PUBLIC_TURNSTILE_SECRET_KEY=0x4AAAAAAABkMYinukE8nzYSZq81xPmL3v`}, nil},
	})
}

func TestCloudflareAccessHeader(t *testing.T) {
	hit := []cfWant{{2, finding.CloudflareConfig, finding.High, finding.Review}}
	runCF(t, []cfCase{
		{"email header trusted", "src/admin.ts", []string{`+export default {`, `+  fetch(req) { const who = req.headers.get("Cf-Access-Authenticated-User-Email"); return ok(who) }`}, hit},
		{"python flask", "app/admin.py", []string{`+def me():`, `+    return request.headers["Cf-Access-Authenticated-User-Email"]`}, hit},
		{"jwt validated in file", "src/admin.ts", []string{`+const jwt = req.headers.get("Cf-Access-Jwt-Assertion")`, `+const who = req.headers.get("Cf-Access-Authenticated-User-Email")`}, nil},
		{"jwt validated in context", "src/admin.ts", []string{` const jwt = req.headers.get("cf-access-jwt-assertion")`, `+const who = req.headers.get("Cf-Access-Authenticated-User-Email")`}, nil},
		{"comment mention", "src/admin.ts", []string{`+// reads Cf-Access-Authenticated-User-Email`}, nil},
		{"markdown doc", "docs/access.md", []string{`+Read Cf-Access-Authenticated-User-Email in the Worker.`}, nil},
	})
}
