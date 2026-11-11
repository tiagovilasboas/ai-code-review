package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseExample(t *testing.T, name string) []File {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "security", "examples", name))
	if err != nil {
		t.Fatal(err)
	}
	return Parse(string(b))
}

func findFile(t *testing.T, files []File, path string) File {
	t.Helper()
	for _, f := range files {
		if f.Path == path {
			return f
		}
	}
	t.Fatalf("missing file %s", path)
	return File{}
}

func addedLine(t *testing.T, files []File, path, snippet string) int {
	t.Helper()
	for _, l := range findFile(t, files, path).Added {
		if strings.Contains(l.Text, snippet) {
			return l.Line
		}
	}
	t.Fatalf("missing added line with %q in %s", snippet, path)
	return 0
}

type lineCase struct {
	path, snippet string
	want          int
}

func checkLines(t *testing.T, files []File, cases []lineCase) {
	t.Helper()
	for _, c := range cases {
		if got := addedLine(t, files, c.path, c.snippet); got != c.want {
			t.Errorf("%s %q: got %d, want %d", c.path, c.snippet, got, c.want)
		}
	}
}

func TestXSSFixture(t *testing.T) {
	files := parseExample(t, "xss-sink.sample.diff")
	checkLines(t, files, []lineCase{
		{"web/src/components/CommentBody.tsx", "dangerouslySetInnerHTML", 18},
	})
}

func TestSSRFFixture(t *testing.T) {
	files := parseExample(t, "ssrf-egress.sample.diff")
	checkLines(t, files, []lineCase{{"src/api/preview.ts", "fetch(targetUrl", 14}})
	for _, l := range findFile(t, files, "src/api/preview.ts").Added {
		if strings.Contains(l.Text, "status.example.invalid") {
			t.Fatal("hardcoded ping must not be an added line")
		}
	}
}

func TestSupplyChainFixture(t *testing.T) {
	files := parseExample(t, "supply-chain.sample.diff")
	checkLines(t, files, []lineCase{
		{"package.json", "postinstall", 8},
		{".npmrc", "registry=", 2},
		{".github/workflows/ci.yml", "some-org/cache-warmup@v1", 9},
		{".github/workflows/ci.yml", "npm install", 14},
	})
}

func TestSamplePRFixture(t *testing.T) {
	files := parseExample(t, "sample-pr.diff")
	checkLines(t, files, []lineCase{
		{"src/api/orders.ts", "findById", 36},
		{"config/app.ts", "paymentApiToken", 14},
		{"src/server.ts", `origin: "*"`, 8},
	})
}

func TestParseMechanics(t *testing.T) {
	text := "diff --git a/x b/x\r\n--- a/x\r\n+++ b/x\r\n@@ -1,2 +5,3 @@\r\n ctx\r\n-old\r\n+new\r\n\\ No newline\r\n" +
		"diff --git a/gone b/gone\n+++ /dev/null\n@@ -1 +0,0 @@\n-bye\n"
	files := Parse(text)
	if len(files) != 1 || files[0].Path != "x" {
		t.Fatalf("files = %+v", files)
	}
	f := files[0]
	if len(f.Lines) != 2 || f.Lines[0].Kind != Context || f.Lines[0].Line != 5 || f.Lines[0].Text != "ctx" {
		t.Fatalf("lines = %+v", f.Lines)
	}
	if len(f.Added) != 1 || f.Added[0].Line != 6 || f.Added[0].Text != "new" || f.Added[0].Kind != Added {
		t.Fatalf("added = %+v", f.Added)
	}
}

func TestParseEmpty(t *testing.T) {
	if got := Parse(""); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseRecordsRemovedLinesAnchoredInsideHunk(t *testing.T) {
	text := "diff --git a/rl.tf b/rl.tf\n--- a/rl.tf\n+++ b/rl.tf\n@@ -1,4 +1,3 @@\n resource \"x\" \"y\" {\n-  phase = \"http_ratelimit\"\n   kind = \"zone\"\n-}"
	files := Parse(text)
	if len(files) != 1 || len(files[0].Removed) != 2 {
		t.Fatalf("files = %+v", files)
	}
	r := files[0].Removed
	if r[0].Kind != Deleted || r[0].Line != 2 || r[0].Text != `  phase = "http_ratelimit"` {
		t.Errorf("first removed = %+v", r[0])
	}
	if r[1].Line != 2 {
		t.Errorf("deletion at hunk end must anchor on the last new-side line, got %d", r[1].Line)
	}
}

func TestParseHunkContentLooksLikeHeaders(t *testing.T) {
	text := "diff --git a/f.txt b/f.txt\n--- a/f.txt\n+++ b/f.txt\n@@ -1,3 +1,4 @@\n keep\n--- a/evil\n+++ b/evil\n+tail\n ctx\n"
	files := Parse(text)
	if len(files) != 1 || files[0].Path != "f.txt" {
		t.Fatalf("files = %+v", files)
	}
	f := files[0]
	if len(f.Added) != 2 || f.Added[0].Text != "++ b/evil" || f.Added[0].Line != 2 || f.Added[1].Text != "tail" || f.Added[1].Line != 3 {
		t.Errorf("added = %+v", f.Added)
	}
	if len(f.Removed) != 1 || f.Removed[0].Text != "-- a/evil" {
		t.Errorf("removed = %+v", f.Removed)
	}
}

func TestParseMultiFileSplitsAfterHunk(t *testing.T) {
	text := "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-x\n+y\n--- a/b.go\n+++ b/b.go\n@@ -1,2 +1,2 @@\n k\n-o\n+n\n"
	files := Parse(text)
	if len(files) != 2 || files[0].Path != "a.go" || files[1].Path != "b.go" {
		t.Fatalf("files = %+v", files)
	}
	if len(files[0].Added) != 1 || files[0].Added[0].Line != 1 || len(files[1].Added) != 1 || files[1].Added[0].Line != 2 {
		t.Errorf("added = %+v / %+v", files[0].Added, files[1].Added)
	}
}

func TestParseHunkHeaderOmittedCounts(t *testing.T) {
	text := "--- a/o.go\n+++ b/o.go\n@@ -1 +1 @@\n-a\n+++ b/evil\n"
	files := Parse(text)
	if len(files) != 1 || files[0].Path != "o.go" || len(files[0].Added) != 1 || files[0].Added[0].Text != "++ b/evil" {
		t.Fatalf("files = %+v", files)
	}
}
