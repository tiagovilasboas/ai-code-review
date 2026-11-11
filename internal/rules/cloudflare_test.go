package rules

import (
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

type cfCase struct {
	name  string
	path  string
	lines []string // "+" added, " " context, "-" removed (anchored on the next new-side line)
	want  []cfWant // empty: must stay silent
}

type cfWant struct {
	line  int
	class finding.Class
	sev   finding.Severity
	act   finding.Action
}

func cfFile(path string, lines []string) diff.File {
	var kept []string
	var removed []diff.Line
	for _, raw := range lines {
		if raw[0] == '-' {
			removed = append(removed, diff.Line{Path: path, Line: len(kept) + 1, Text: raw[1:], Kind: diff.Deleted})
			continue
		}
		kept = append(kept, raw)
	}
	f := mkFile(path, kept...)
	f.Removed = removed
	return f
}

func runCF(t *testing.T, cases []cfCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Cloudflare(cfFile(c.path, c.lines))
			if len(got) != len(c.want) {
				t.Fatalf("got %d findings %+v, want %d", len(got), got, len(c.want))
			}
			for i, w := range c.want {
				g := got[i]
				if g.Line != w.line || g.Class != w.class || g.Severity != w.sev || g.Action != w.act {
					t.Fatalf("finding %d = %+v, want %+v", i, g, w)
				}
			}
		})
	}
}

func TestCloudflareFindingsAreEvidenceBacked(t *testing.T) {
	got := Cloudflare(cfFile("wrangler.toml", []string{"+[vars]", `+STRIPE_SECRET_KEY = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`}))
	if len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
	g := got[0]
	if g.Path != "wrangler.toml" || g.Line != 2 || g.CWE != "CWE-798" || g.Action != finding.Block {
		t.Fatalf("finding = %+v", g)
	}
	if strings.Contains(g.Title, "k3Jx9") {
		t.Fatalf("title leaks the value: %q", g.Title)
	}
}

func TestCloudflareOWASPIDsAreOfficial(t *testing.T) {
	files := []diff.File{
		cfFile("zone.tf", []string{` resource "cloudflare_zone_settings_override" "z" {`, `+  ssl = "off"`}),
		cfFile("rl.tf", []string{` resource "cloudflare_ruleset" "rl" {`, `+  enabled = false`, `   phase = "http_ratelimit"`}),
		cfFile("src/a.ts", []string{`+const e = req.headers.get("Cf-Access-Authenticated-User-Email")`}),
		cfFile("wrangler.toml", []string{`+workers_dev = true`}),
	}
	for _, f := range files {
		for _, g := range Cloudflare(f) {
			for _, id := range g.OWASP {
				if !finding.IsOfficialID(id) {
					t.Errorf("%s: unofficial id %s", f.Path, id)
				}
			}
		}
	}
}

func TestCloudflareSkipsTestPaths(t *testing.T) {
	runCF(t, []cfCase{
		{"test dir wrangler", "tests/wrangler.toml", []string{"+[vars]", `+API_TOKEN = "k3Jx9Qm2Lp7Vw5Rt8Yu1"`}, nil},
		{"spec file", "src/access.spec.ts", []string{`+const e = req.headers.get("cf-access-authenticated-user-email")`}, nil},
	})
}
