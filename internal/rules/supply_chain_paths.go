package rules

import (
	"path"
	"strings"
)

var testDirs = map[string]bool{
	"test": true, "tests": true, "__tests__": true, "spec": true, "specs": true,
	"fixtures": true, "fixture": true, "__fixtures__": true, "testdata": true, "__mocks__": true,
}

// isTestPath reports whether p is test data or a test file: such content is
// quoted or copied on purpose and is never installed or followed.
func isTestPath(p string) bool {
	dir, base := path.Split(strings.ToLower(p))
	for _, seg := range strings.Split(dir, "/") {
		if testDirs[seg] {
			return true
		}
	}
	return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") ||
		strings.Contains(base, "_test.") || strings.HasPrefix(base, "test_")
}
