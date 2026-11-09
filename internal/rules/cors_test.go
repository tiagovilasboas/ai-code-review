package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestCORSShapes(t *testing.T) {
	cases := []struct {
		name string
		body []string
		want int
	}{
		{"cors() wildcard", []string{"+app.use(cors({ origin: '*' }));"}, 1},
		{"cors() wildcard credentials", []string{"+app.use(cors({ origin: '*', credentials: true }));"}, 1},
		{"read-only wildcard without credentials", []string{"+const c = cors({ origin: '*', methods: ['GET'] });"}, 0},
		{"static asset wildcard", []string{"+app.use('/assets', cors({ origin: '*' }), express.static('p'));"}, 0},
		{"manual header wildcard", []string{"+res.setHeader('Access-Control-Allow-Origin', '*');"}, 1},
		{"manual header wildcard with write methods", []string{
			"+res.setHeader('Access-Control-Allow-Origin', '*');",
			"+res.setHeader('Access-Control-Allow-Methods', 'GET,POST');"}, 1},
		{"manual header wildcard GET only", []string{
			"+res.setHeader('Access-Control-Allow-Origin', '*');",
			"+res.setHeader('Access-Control-Allow-Methods', 'GET');"}, 0},
		{"nginx wildcard", []string{"+add_header Access-Control-Allow-Origin *;"}, 1},
		{"origin true with credentials", []string{"+register(cors, { origin: true,", "+  credentials: true });"}, 1},
		{"origin true without credentials", []string{"+register(cors, { origin: true });"}, 0},
		{"reflect header with credentials", []string{
			"+res.setHeader('Access-Control-Allow-Origin', req.headers.origin);",
			"+res.setHeader('Access-Control-Allow-Credentials', 'true');"}, 1},
		{"allowlisted origin value", []string{
			"+res.setHeader('Access-Control-Allow-Origin', allowed);",
			"+res.setHeader('Access-Control-Allow-Credentials', 'true');"}, 0},
		{"commented wildcard", []string{"+// res.setHeader('Access-Control-Allow-Origin', '*');"}, 0},
		{"origin star outside cors context", []string{"+const opts = { origin: '*' };"}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(Secrets(fileOf("src/s.js", 1, c.body...)), finding.CORS)
			if got != c.want {
				t.Fatalf("got %d cors findings, want %d", got, c.want)
			}
		})
	}
}
