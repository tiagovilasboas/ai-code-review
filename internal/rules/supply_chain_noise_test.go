package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// withRemoved returns f with deleted lines anchored on the given new-side line.
func withRemoved(f diff.File, anchor int, texts ...string) diff.File {
	for _, t := range texts {
		f.Removed = append(f.Removed, diff.Line{Path: f.Path, Line: anchor, Text: t, Kind: diff.Deleted})
	}
	return f
}

func TestLockfileIntegrityComparesValues(t *testing.T) {
	const h = "sha512-AAAA=="
	cases := []struct {
		name    string
		file    diff.File
		wantHit bool
	}{
		{"npm trailing comma and peer flag only",
			withRemoved(mkFile("package-lock.json",
				` "node_modules/ms": {`, `   "version": "1.3.0",`,
				`   "resolved": "https://registry.npmjs.org/ms/-/ms-1.3.0.tgz",`,
				`+  "integrity": "`+h+`",`, `+  "peer": true`, ` }`),
				4, `  "integrity": "`+h+`"`),
			false},
		{"npm entry moved elsewhere in the file",
			withRemoved(mkFile("package-lock.json",
				` "node_modules/ms": {`, `+  "version": "1.3.0",`,
				`+  "resolved": "https://registry.npmjs.org/ms/-/ms-1.3.0.tgz",`,
				`+  "integrity": "`+h+`"`),
				1, `  "version": "1.3.0",`, `  "resolved": "https://registry.npmjs.org/ms/-/ms-1.3.0.tgz",`, `  "integrity": "`+h+`"`),
			false},
		{"yarn same hash re-added",
			withRemoved(mkFile("yarn.lock", ` ms@^1.3.0:`, `   version "1.3.0"`, `+  integrity `+h),
				3, `  integrity `+h),
			false},
		{"npm real swap on the same version",
			withRemoved(mkFile("package-lock.json",
				` "node_modules/ms": {`, `   "version": "1.3.0",`,
				`   "resolved": "https://registry.npmjs.org/ms/-/ms-1.3.0.tgz",`,
				`+  "integrity": "sha512-BBBB==",`, ` }`),
				4, `  "integrity": "`+h+`",`),
			true},
		{"npm swap, other package keeps the old hash removed",
			withRemoved(mkFile("package-lock.json",
				` "node_modules/ms": {`, `   "version": "1.3.0",`, `+  "integrity": "sha512-BBBB=="`),
				3, `  "integrity": "sha512-CCCC=="`),
			true},
	}
	for _, c := range cases {
		got := SupplyChain(c.file)
		if (len(got) > 0) != c.wantHit {
			t.Errorf("%s: wantHit=%v got %+v", c.name, c.wantHit, got)
		}
	}
}
