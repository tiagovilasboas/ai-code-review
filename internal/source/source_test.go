package source

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	gitT(t, dir, "config", "user.name", "T")
	gitT(t, dir, "config", "user.email", "t@example.com")
	gitT(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", msg)
}

// hasAdded asserts the parsed diff carries an added line at path:line.
func hasAdded(t *testing.T, text, path string, line int, want string) {
	t.Helper()
	for _, f := range diff.Parse(text) {
		for _, l := range f.Added {
			if f.Path == path && l.Line == line && l.Text == want {
				return
			}
		}
	}
	t.Fatalf("no added %s:%d %q in:\n%s", path, line, want, text)
}

func TestWorkspaceStagedUnstagedUntracked(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.txt", "one\ntwo\n")
	write(t, dir, "b.txt", "x\n")
	write(t, dir, "gone.txt", "bye\n")
	commit(t, dir, "init")

	write(t, dir, "a.txt", "one\ntwo\nstaged\n")
	gitT(t, dir, "add", "a.txt")
	write(t, dir, "b.txt", "x\nunstaged\n")
	write(t, dir, "sub/new.txt", "fresh\nline\n")
	write(t, dir, "bin.dat", "ab\x00cd")
	write(t, dir, ".gitignore", "ignored.txt\n")
	write(t, dir, "ignored.txt", "nope\n")
	if err := os.Remove(filepath.Join(dir, "gone.txt")); err != nil {
		t.Fatal(err)
	}

	out, err := Workspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	hasAdded(t, out, "a.txt", 3, "staged")
	hasAdded(t, out, "b.txt", 2, "unstaged")
	hasAdded(t, out, "sub/new.txt", 2, "line")
	if !strings.Contains(out, "diff --git a/gone.txt b/gone.txt") {
		t.Errorf("deletion missing:\n%s", out)
	}
	for _, bad := range []string{"bin.dat", "ignored.txt"} {
		if strings.Contains(out, "b/"+bad) {
			t.Errorf("%s must not appear:\n%s", bad, out)
		}
	}
}

func TestWorkspaceFromSubdirAndNoCommits(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "s/staged.txt", "hello\n")
	gitT(t, dir, "add", "s/staged.txt")
	write(t, dir, "loose.txt", "loose\n")

	out, err := Workspace(filepath.Join(dir, "s"))
	if err != nil {
		t.Fatal(err)
	}
	hasAdded(t, out, "s/staged.txt", 1, "hello")
	hasAdded(t, out, "loose.txt", 1, "loose")
}

func TestRangeUsesMergeBase(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "base.txt", "base\n")
	commit(t, dir, "base")
	gitT(t, dir, "switch", "-q", "-c", "feature")
	write(t, dir, "feat.txt", "f1\nf2\n")
	commit(t, dir, "feature")
	gitT(t, dir, "switch", "-q", "main")
	write(t, dir, "main-only.txt", "m\n")
	commit(t, dir, "main moves")

	out, err := Range(dir, "main", "feature")
	if err != nil {
		t.Fatal(err)
	}
	hasAdded(t, out, "feat.txt", 2, "f2")
	if strings.Contains(out, "main-only.txt") {
		t.Errorf("merge-base semantics violated:\n%s", out)
	}
}

func TestCommitAndRootCommit(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "r.txt", "root\n")
	commit(t, dir, "root")
	root := gitT(t, dir, "rev-parse", "HEAD")
	write(t, dir, "r.txt", "root\nsecond\n")
	commit(t, dir, "second")
	second := gitT(t, dir, "rev-parse", "HEAD")

	out, err := Commit(dir, second)
	if err != nil {
		t.Fatal(err)
	}
	hasAdded(t, out, "r.txt", 2, "second")
	if strings.Contains(out, "+root") {
		t.Errorf("parent content leaked:\n%s", out)
	}

	out, err = Commit(dir, root)
	if err != nil {
		t.Fatal(err)
	}
	hasAdded(t, out, "r.txt", 1, "root")
}

func TestPrefixConfigCannotChangePaths(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "p.txt", "a\n")
	commit(t, dir, "init")
	gitT(t, dir, "config", "diff.noprefix", "true")
	gitT(t, dir, "config", "diff.mnemonicPrefix", "true")
	write(t, dir, "p.txt", "a\nb\n")
	out, err := Workspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	hasAdded(t, out, "p.txt", 2, "b")
}

func TestRejectsOptionLikeRefs(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "x.txt", "x\n")
	commit(t, dir, "init")
	if _, err := Range(dir, "--output=/tmp/pwn", "HEAD"); err == nil || !strings.Contains(err.Error(), "'-'") {
		t.Errorf("range from: %v", err)
	}
	if _, err := Range(dir, "HEAD", "-x"); err == nil {
		t.Error("range to: want error")
	}
	if _, err := Commit(dir, "--help"); err == nil {
		t.Error("commit: want error")
	}
	if _, err := Commit(dir, "nosuchref"); err == nil {
		t.Error("unknown commit: want error")
	}
}

func TestNotARepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	_, err := Workspace(dir)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "not a git repository") {
		t.Errorf("want stderr in error, got %v", err)
	}
}

func TestFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "d.diff", "x")
	got, err := File(filepath.Join(dir, "d.diff"))
	if err != nil || got != "x" {
		t.Errorf("got %q, %v", got, err)
	}
}
