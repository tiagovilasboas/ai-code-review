package source

import (
	"strings"
)

// workspace implements Workspace; see the exported doc in source.go.
func workspace(dir string) (string, error) {
	top, err := runGit(dir, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	top = strings.TrimSpace(top)

	base := "HEAD"
	if !hasCommit(top, "HEAD") {
		if base, err = emptyTree(top); err != nil {
			return "", err
		}
	}
	tracked, err := gitDiff(top, base)
	if err != nil {
		return "", err
	}
	untracked, err := untrackedDiff(top)
	if err != nil {
		return "", err
	}
	return tracked + untracked, nil
}

// untrackedDiff renders each untracked, non-ignored file as a new-file diff.
func untrackedDiff(top string) (string, error) {
	list, err := runGit(top, nil, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, f := range strings.Split(list, "\x00") {
		if f == "" {
			continue
		}
		args := append([]string{"diff", "--no-index"}, diffFlags...)
		args = append(args, "--", "/dev/null", f)
		out, err := runGit(top, []int{1}, args...)
		if err != nil {
			return "", err
		}
		b.WriteString(dropBinary(normalizeNewFile(out, f)))
	}
	return b.String(), nil
}

// normalizeNewFile forces the header paths to a/<f> and b/<f>.
func normalizeNewFile(out, f string) string {
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		switch {
		case strings.HasPrefix(l, "diff --git "):
			lines[i] = "diff --git a/" + f + " b/" + f
		case strings.HasPrefix(l, "+++ ") && l != "+++ /dev/null":
			lines[i] = "+++ b/" + f
		}
	}
	return strings.Join(lines, "\n")
}
