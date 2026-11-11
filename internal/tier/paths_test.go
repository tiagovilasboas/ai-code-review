package tier

import (
	"reflect"
	"testing"
)

func TestChangedPaths(t *testing.T) {
	cases := []struct {
		name, diff string
		want       []string
	}{
		{"empty", "", []string{}},
		{"modify", "diff --git a/x.go b/x.go\n--- a/x.go\n+++ b/x.go\n@@ -1 +1 @@\n-a\n+b\n", []string{"x.go"}},
		{"delete", "diff --git a/old.go b/old.go\ndeleted file mode 100644\n--- a/old.go\n+++ /dev/null\n@@ -1 +0,0 @@\n-a\n", []string{"old.go"}},
		{"add", "diff --git a/n.go b/n.go\nnew file mode 100644\n--- /dev/null\n+++ b/n.go\n@@ -0,0 +1 @@\n+a\n", []string{"n.go"}},
		{"pure rename", "diff --git a/a.md b/docs/b.md\nsimilarity index 100%\nrename from a.md\nrename to docs/b.md\n", []string{"a.md", "docs/b.md"}},
		{"mode only", "diff --git a/run.sh b/run.sh\nold mode 100644\nnew mode 100755\n", []string{"run.sh"}},
		{"spaces", "diff --git a/my file.md b/my file.md\nold mode 100644\nnew mode 100755\n", []string{"my file.md"}},
		{"plain unified", "--- a/p.go\t2026-01-01\n+++ b/p.go\t2026-01-02\n@@ -1 +1 @@\n-a\n+b\n", []string{"p.go"}},
		{"crlf", "diff --git a/c.go b/c.go\r\n--- a/c.go\r\n+++ b/c.go\r\n", []string{"c.go"}},
	}
	for _, c := range cases {
		got := ChangedPaths(c.diff)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
