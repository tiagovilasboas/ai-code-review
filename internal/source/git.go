package source

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// diffFlags pin the output format so user config cannot change what the
// parser sees (colors, external drivers, textconv, path prefixes).
var diffFlags = []string{
	"--no-color", "--no-ext-diff", "--no-textconv",
	"--src-prefix=a/", "--dst-prefix=b/", "-U3",
}

// runGit runs git without a shell, in dir, and returns stdout. Exit codes in
// okCodes are not errors (git diff --no-index exits 1 when files differ).
func runGit(dir string, okCodes []int, args ...string) (string, error) {
	full := append([]string{"-c", "core.quotepath=off"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			for _, c := range okCodes {
				if ee.ExitCode() == c {
					return out.String(), nil
				}
			}
		}
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], msg)
	}
	return out.String(), nil
}

// checkRef rejects values git could read as an option.
func checkRef(kind, ref string) error {
	if ref == "" {
		return fmt.Errorf("%s must not be empty", kind)
	}
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("invalid %s %q: must not start with '-'", kind, ref)
	}
	return nil
}

func emptyTree(dir string) (string, error) {
	out, err := runGit(dir, nil, "hash-object", "-t", "tree", "/dev/null")
	return strings.TrimSpace(out), err
}

func hasCommit(dir, rev string) bool {
	_, err := runGit(dir, nil, "rev-parse", "--verify", "--quiet", rev+"^{commit}")
	return err == nil
}

func gitDiff(dir string, revs ...string) (string, error) {
	args := append([]string{"diff"}, diffFlags...)
	args = append(args, revs...)
	args = append(args, "--")
	out, err := runGit(dir, nil, args...)
	if err != nil {
		return "", err
	}
	return dropBinary(out), nil
}

// dropBinary removes per-file blocks git could not render as text hunks.
func dropBinary(diff string) string {
	const marker = "diff --git "
	if !strings.Contains(diff, marker) {
		return diff
	}
	var b strings.Builder
	for _, blk := range splitBlocks(diff) {
		if strings.Contains(blk, "\nBinary files ") || strings.Contains(blk, "\nGIT binary patch") {
			continue
		}
		b.WriteString(blk)
	}
	return b.String()
}

// splitBlocks splits a diff into blocks that each start at "diff --git ".
func splitBlocks(diff string) []string {
	var blocks []string
	start := 0
	for i := 1; i < len(diff); i++ {
		if diff[i-1] == '\n' && strings.HasPrefix(diff[i:], "diff --git ") {
			blocks = append(blocks, diff[start:i])
			start = i
		}
	}
	return append(blocks, diff[start:])
}
