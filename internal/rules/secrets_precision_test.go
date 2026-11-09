package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func countClass(fs []finding.Finding, c finding.Class) int {
	n := 0
	for _, f := range fs {
		if f.Class == c {
			n++
		}
	}
	return n
}

func TestSecretsCredentialShapes(t *testing.T) {
	cases := []struct {
		name, path, line string
		want             int
	}{
		{"db url literal password", "src/a.js", `+const u = 'postgres://app:Zq81xPmL@db.internal:5432/app';`, 1},
		{"db url interpolated", "src/a.js", "+const u = `postgres://app:${pw}@db.internal/app`;", 0},
		{"db url placeholder", "src/a.js", `+const u = 'mysql://user:<password>@db/app';`, 0},
		{"db url local dev pair", "src/a.js", `+const u = 'postgres://dev:hunter22@localhost/app';`, 0},
		{"db url without password", "src/a.js", `+const u = 'postgres://app@db.internal/app';`, 0},
		{"pem private key", "src/k.ts", "+-----BEGIN EC PRIVATE KEY-----", 1},
		{"pem public key", "src/k.ts", "+-----BEGIN PUBLIC KEY-----", 0},
		{"npmrc literal token", ".npmrc", "+//registry.example.org/:_authToken=abc123def456", 1},
		{"npmrc env token", ".npmrc", "+//registry.example.org/:_authToken=${NPM_TOKEN}", 0},
		{"slashes comment mentioning token", "src/a.js", "+// _authToken=abc123def456", 0},
		{"cookie name constant", "src/c.ts", `+  token: 'session-token',`, 0},
		{"header name constant", "src/c.ts", `+  apiKey: 'x-api-key',`, 0},
		{"dummy password", "src/c.ts", `+  password: 'changeme',`, 0},
		{"literal password in source", "src/c.ts", `+  password: 'Tr0ub4dor&3',`, 1},
		{"literal password in test file", "src/c.test.ts", `+  password: 'Tr0ub4dor&3',`, 0},
		{"known prefix in test file still flagged", "src/c.test.ts", `+  const k = 'sk_live_EXAMPLEEXAMPLE00';`, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(Secrets(fileOf(c.path, 1, c.line)), finding.Secret)
			if got != c.want {
				t.Fatalf("got %d secret findings, want %d", got, c.want)
			}
		})
	}
}

func TestSecretProseLabelIsSilent(t *testing.T) {
	f := fileOf("internal/finding/finding.go", 1, `+	Secret:          "Secret in source",`, `+	token: "s3cr3t-Value-91x",`)
	got := Secrets(f)
	if len(got) != 1 || got[0].Line != 2 {
		t.Fatalf("only the credential-shaped literal must fire, got %+v", got)
	}
}
