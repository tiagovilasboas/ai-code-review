package rules

import (
	"path"
	"regexp"
	"strings"
)

// camelTestFile matches PHPUnit and JUnit class files (OrderTest.php,
// OrderServiceTests.java); the capital T keeps latest.php out.
var camelTestFile = regexp.MustCompile(`[a-z0-9]Tests?\.(?:php|java|kt|scala)$`)

var testDirs = map[string]bool{
	"test": true, "tests": true, "__tests__": true, "spec": true, "specs": true,
	"fixtures": true, "fixture": true, "e2e": true, "__fixtures__": true, "testdata": true, "__mocks__": true,
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
	if camelTestFile.MatchString(p) {
		return true
	}
	return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") ||
		strings.Contains(base, "_test.") || strings.HasPrefix(base, "test_")
}
