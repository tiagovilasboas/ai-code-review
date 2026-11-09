// Package source produces the unified diff the review engine reads: from a
// file, or straight from git (workspace, a ref range, or one commit).
//
// Git sources shell out to the `git` binary (no shell, args as a slice) and
// need git 2.x with `diff --src-prefix`; tested with 2.39.
package source

import (
	"fmt"
	"os"
)

// File reads a unified diff from path.
func File(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

// Workspace returns staged, unstaged and untracked changes of the repo at dir
// against HEAD, as one unified diff with b/ paths. A repo without commits is
// diffed against the empty tree. Binary files are skipped.
func Workspace(dir string) (string, error) { return workspace(dir) }

// Range returns the changes of `to` since it diverged from `from`
// (merge-base semantics, like `git diff from...to`).
func Range(dir, from, to string) (string, error) {
	if err := checkRef("ref", from); err != nil {
		return "", err
	}
	if err := checkRef("ref", to); err != nil {
		return "", err
	}
	return gitDiff(dir, from+"..."+to)
}

// Commit returns the changes introduced by one commit against its first parent;
// a root commit is diffed against the empty tree.
func Commit(dir, sha string) (string, error) {
	if err := checkRef("commit", sha); err != nil {
		return "", err
	}
	if !hasCommit(dir, sha) {
		return "", fmt.Errorf("git: unknown commit %q", sha)
	}
	if !hasCommit(dir, sha+"^") {
		empty, err := emptyTree(dir)
		if err != nil {
			return "", err
		}
		return gitDiff(dir, empty, sha)
	}
	return gitDiff(dir, sha+"^", sha)
}
