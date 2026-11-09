// Package source produces the unified diff the review engine reads: from a
// file, or straight from git (workspace, a ref range, or one commit).
//
// Contract only: the git-backed functions are implemented in a follow-up.
// Callers must treat ErrNotImplemented like any other source error.
package source

import (
	"errors"
	"os"
)

// ErrNotImplemented is returned by git sources until they are implemented.
var ErrNotImplemented = errors.New("git source not implemented")

// File reads a unified diff from path.
func File(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

// Workspace returns staged, unstaged and untracked changes of the repo at dir
// against HEAD, as one unified diff with b/ paths.
func Workspace(dir string) (string, error) { return "", ErrNotImplemented }

// Range returns the changes of `to` since it diverged from `from`
// (merge-base semantics, like `git diff from...to`).
func Range(dir, from, to string) (string, error) { return "", ErrNotImplemented }

// Commit returns the changes introduced by one commit against its first parent.
func Commit(dir, sha string) (string, error) { return "", ErrNotImplemented }
