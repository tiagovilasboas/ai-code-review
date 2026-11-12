package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestCORSGoFiberWildcard(t *testing.T) {
	runRuleCases(t, corsFindings, finding.CORS, "CWE-942", []ruleCase{
		{"fiber wildcard with credentials", mkFile("cmd/api/main.go",
			`+app.Use(cors.New(cors.Config{`,
			`+	AllowOrigins:     "*",`,
			`+	AllowCredentials: true,`,
			`+}))`,
		), []int{2}, finding.Medium},
		{"fiber wildcard on one line", mkFile("cmd/api/main.go",
			`+app.Use(cors.New(cors.Config{AllowOrigins: "*", AllowCredentials: true}))`,
		), []int{1}, finding.Medium},
		{"fiber wildcard inside a list", mkFile("cmd/api/main.go",
			`+app.Use(cors.New(cors.Config{AllowOrigins: "https://a.example.com, *", AllowCredentials: true}))`,
		), []int{1}, finding.Medium},
		{"fiber explicit origins", mkFile("cmd/api/main.go",
			`+app.Use(cors.New(cors.Config{AllowOrigins: "https://a.example.com, https://b.example.com", AllowCredentials: true}))`,
		), nil, finding.Medium},
		{"fiber subdomain pattern is not a bare wildcard", mkFile("cmd/api/main.go",
			`+app.Use(cors.New(cors.Config{AllowOrigins: "https://*.example.com", AllowCredentials: true}))`,
		), nil, finding.Medium},
	})
}
