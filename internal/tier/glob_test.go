package tier

import "testing"

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		glob, path string
		want       bool
	}{
		{"docs/**", "docs/a.md", true},
		{"docs/**", "docs/x/y/z.md", true},
		{"docs/**", "docsx/a.md", false},
		{"docs/**", "a/docs/a.md", false},
		{"**/*.go", "main.go", true},
		{"**/*.go", "a/b/main.go", true},
		{"**/*.go", "a/b/main.js", false},
		{"internal/**/catalog.go", "internal/finding/catalog.go", true},
		{"internal/**/catalog.go", "internal/a/b/catalog.go", true},
		{"internal/**/catalog.go", "internal/catalog.go", true},
		{"*.md", "README.md", true},
		{"*.md", "docs/deep/a.md", true},
		{"*.md", "a.mdx", false},
		{"action.yml", "action.yml", true},
		{"action.yml", "sub/action.yml", true},
		{"scripts/install-review-process.*", "scripts/install-review-process.sh", true},
		{"scripts/install-review-process.*", "scripts/install-review-process.ps1", true},
		{".github/workflows/**", ".github/workflows/ci.yml", true},
		{".github/workflows/**", ".github/dependabot.yml", false},
		{"src/*/x.js", "src/a/x.js", true},
		{"src/*/x.js", "src/a/b/x.js", false},
		{"**", "any/thing", true},
	}
	for _, c := range cases {
		if got := matchGlob(c.glob, c.path); got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.glob, c.path, got, c.want)
		}
	}
}

func TestValidGlob(t *testing.T) {
	for _, g := range []string{"", "  ", "a**b", "src/[x"} {
		if validGlob(g) == nil {
			t.Errorf("validGlob(%q) accepted", g)
		}
	}
	for _, g := range []string{"**", "a/**/b", "*.md", "x/?.go"} {
		if err := validGlob(g); err != nil {
			t.Errorf("validGlob(%q): %v", g, err)
		}
	}
}
