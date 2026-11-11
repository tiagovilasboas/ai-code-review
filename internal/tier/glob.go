package tier

import (
	"fmt"
	"path"
	"strings"
)

// Glob syntax (slash-separated paths, no leading slash):
//
//	*   any run of characters within one path segment (path.Match rules)
//	?   one character within a segment
//	**  a whole segment that matches zero or more segments
//
// A pattern with no "/" matches the base name at any depth, like .gitignore
// ("*.md" matches docs/a.md). A pattern ending in "/**" matches everything
// below that directory.

func validGlob(g string) error {
	if strings.TrimSpace(g) == "" {
		return fmt.Errorf("empty glob")
	}
	for _, seg := range strings.Split(g, "/") {
		if seg == "**" {
			continue
		}
		if strings.Contains(seg, "**") {
			return fmt.Errorf("glob %q: ** must be a whole path segment", g)
		}
		if _, err := path.Match(seg, ""); err != nil {
			return fmt.Errorf("glob %q: %w", g, err)
		}
	}
	return nil
}

// matchGlob reports whether the slash-separated file path matches glob.
func matchGlob(glob, file string) bool {
	if !strings.Contains(glob, "/") {
		ok, _ := path.Match(glob, path.Base(file))
		return ok
	}
	return matchSegs(strings.Split(glob, "/"), strings.Split(file, "/"))
}

func matchSegs(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			if len(rest) == 0 {
				return len(segs) > 0
			}
			for i := 0; i <= len(segs); i++ {
				if matchSegs(rest, segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		if ok, _ := path.Match(pat[0], segs[0]); !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

// firstMatch returns the first glob in globs matching file.
func firstMatch(globs []string, file string) (string, bool) {
	for _, g := range globs {
		if matchGlob(g, file) {
			return g, true
		}
	}
	return "", false
}
