package tier

import (
	"sort"
	"strconv"
	"strings"
)

// ChangedPaths lists every path a unified diff touches: new paths, deleted
// paths, and both sides of a rename or copy. internal/diff only yields the new
// side with hunk lines, which would hide deletions and pure renames, so the
// classifier reads the headers itself. The result is sorted and unique.
func ChangedPaths(diff string) []string {
	set := map[string]bool{}
	add := func(p string) {
		if p = cleanPath(p); p != "" {
			set[p] = true
		}
	}
	for _, line := range strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			a, b := splitGitHeader(line[len("diff --git "):])
			add(a)
			add(b)
		case strings.HasPrefix(line, "--- "), strings.HasPrefix(line, "+++ "):
			add(line[4:])
		case strings.HasPrefix(line, "rename from "):
			add(line[len("rename from "):])
		case strings.HasPrefix(line, "rename to "):
			add(line[len("rename to "):])
		case strings.HasPrefix(line, "copy from "):
			add(line[len("copy from "):])
		case strings.HasPrefix(line, "copy to "):
			add(line[len("copy to "):])
		}
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// splitGitHeader splits `a/X b/Y`. Paths may hold spaces, so it prefers the
// split where both halves name the same file and falls back to the first one.
func splitGitHeader(rest string) (string, string) {
	first := -1
	for i := 0; i+3 <= len(rest); i++ {
		if rest[i] != ' ' || !strings.HasPrefix(rest[i+1:], "b/") {
			continue
		}
		left, right := rest[:i], rest[i+1:]
		if first < 0 {
			first = i
		}
		if strings.TrimPrefix(left, "a/") == strings.TrimPrefix(right, "b/") {
			return left, right
		}
	}
	if first < 0 {
		return rest, ""
	}
	return rest[:first], rest[first+1:]
}

// cleanPath normalizes a diff path: drops timestamps, quotes and the a/ b/
// prefix, and returns "" for /dev/null.
func cleanPath(p string) string {
	p, _, _ = strings.Cut(p, "\t")
	p = strings.TrimSpace(p)
	if strings.HasPrefix(p, `"`) {
		if u, err := strconv.Unquote(p); err == nil {
			p = u
		}
	}
	if p == "" || p == "/dev/null" {
		return ""
	}
	p = strings.TrimPrefix(strings.TrimPrefix(p, "a/"), "b/")
	return strings.TrimPrefix(p, "./")
}
