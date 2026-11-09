package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestCORSReflectedOriginAndSpring(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		{"go inline echo", "a.go", []string{`+w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))`}, 1},
		{"go echo through local", "a.go", []string{
			`+origin := r.Header.Get("Origin")`,
			`+w.Header().Set("Access-Control-Allow-Origin", origin)`}, 1},
		{"go map allowlist", "a.go", []string{
			`+origin := r.Header.Get("Origin")`,
			`+if allowed[origin] {`,
			`+	w.Header().Set("Access-Control-Allow-Origin", origin)`,
			`+}`}, 0},
		{"go slices.Contains allowlist", "a.go", []string{
			`+origin := r.Header.Get("Origin")`,
			`+if slices.Contains(origins, origin) {`,
			`+	w.Header().Set("Access-Control-Allow-Origin", origin)`,
			`+}`}, 0},
		{"koa ctx.get echo", "a.js", []string{`+ctx.set('Access-Control-Allow-Origin', ctx.get('Origin'));`}, 1},
		{"koa request.header.origin", "a.js", []string{`+ctx.set('Access-Control-Allow-Origin', ctx.request.header.origin);`}, 1},
		{"koa includes allowlist", "a.js", []string{
			`+const o = ctx.get('Origin');`,
			`+if (list.includes(o)) ctx.set('Access-Control-Allow-Origin', o);`}, 0},
		{"fixed origin", "a.go", []string{`+w.Header().Set("Access-Control-Allow-Origin", "https://app.example.com")`}, 0},
		{"spring originPatterns with credentials", "A.java", []string{`+@CrossOrigin(originPatterns = "*", allowCredentials = "true")`}, 1},
		{"spring allowedOriginPatterns with credentials", "A.java", []string{
			`+registry.addMapping("/**").allowedOriginPatterns("*")`,
			`+  .allowCredentials(true);`}, 1},
		{"spring origins star", "A.java", []string{`+@CrossOrigin(origins = "*")`}, 1},
		{"spring origins star read only", "A.java", []string{
			`+@CrossOrigin(origins = "*", methods = RequestMethod.GET)`}, 0},
		{"spring explicit origin", "A.java", []string{`+@CrossOrigin(origins = "https://a.example.com", allowCredentials = "true")`}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(Secrets(fileOf("src/"+c.path, 1, c.body...)), finding.CORS)
			if got != c.want {
				t.Fatalf("got %d cors findings, want %d", got, c.want)
			}
		})
	}
}
