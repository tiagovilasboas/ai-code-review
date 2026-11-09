package rules

import (
	"reflect"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestSecretsSamplePR(t *testing.T) {
	cfg := fileOf("config/app.ts", 10,
		"   port: Number(process.env.PORT ?? 3000),",
		"   dbUrl: process.env.DATABASE_URL,",
		"+  // Secrets smoke target: hardcoded token (placeholder shape — not a live key)",
		"+  // Report location only; do not echo the full value in agent output",
		`+  paymentApiToken: "pk_test_REDACTED_SAMPLE_ONLY_0000",`,
		" };",
	)
	got := Secrets(cfg)
	want := []finding.Finding{{
		Title: "Hardcoded credential in source", Class: finding.Secret,
		Path: "config/app.ts", Line: 14, CWE: "CWE-798",
		OWASP:    []string{"A02:2021", "ASVS-5.0-13.3.1"},
		Severity: finding.Critical, Action: finding.Block,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}

	srv := fileOf("src/server.ts", 1,
		` import express from "express";`,
		` import { listOrders, createOrder, getOrder } from "./api/orders";`,
		" ",
		" const app = express();",
		" app.use(express.json());",
		"+",
		"+// Misconfiguration: open CORS introduced in this PR",
		`+app.use(cors({ origin: "*" }));`,
	)
	got = Secrets(srv)
	want = []finding.Finding{{
		Title: "Overly permissive CORS origin", Class: finding.CORS,
		Path: "src/server.ts", Line: 8, CWE: "CWE-942",
		OWASP:    []string{"A05:2021", "ASVS-5.0-3.4.2"},
		Severity: finding.Medium, Action: finding.Review,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestSecrets(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{"password assignment", `+const password = "hunter2";`, []string{"CWE-798"}},
		{"case-insensitive key", `+  API_KEY: 'abc123',`, []string{"CWE-798"}},
		{"known prefix", `+const k = "AKIAABCDEFGHIJKLMNOP";`, []string{"CWE-798"}},
		{"github token prefix", `+call('ghp_0000');`, []string{"CWE-798"}},
		{"process.env is silence", `+const token = process.env.TOKEN || "x";`, nil},
		{"comment is silence", `+// token = "abc"`, nil},
		{"empty value is silence", `+const secret = "";`, nil},
		{"cors star", `+app.use(cors({ origin: '*' }));`, []string{"CWE-942"}},
		{"cors specific origin", `+app.use(cors({ origin: "https://a.example" }));`, nil},
		{"both on one line", `+cors({ origin: "*", token: "abc" })`, []string{"CWE-798", "CWE-942"}},
		{"context line is not evidence", ` const token = "abc";`, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			for _, f := range Secrets(fileOf("a.ts", 1, c.line)) {
				got = append(got, f.CWE)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("CWEs = %v, want %v", got, c.want)
			}
		})
	}
}

func TestSecretsQualifiedMapKeyIsSilent(t *testing.T) {
	f := fileOf("internal/report/sarif.go", 20,
		` var skillFor = map[finding.Class]string{`,
		`+	finding.Secret:          "secrets-config",`,
		`+	cfg.Token: "label",`,
	)
	if got := Secrets(f); len(got) != 0 {
		t.Fatalf("qualified map key must stay silent, got %+v", got)
	}
	f = fileOf("config/app.ts", 1, `+  password: "hunter22",`, `+  db.password = "hunter22";`)
	if got := Secrets(f); len(got) != 2 {
		t.Fatalf("plain keys must still fire twice, got %+v", got)
	}
}
