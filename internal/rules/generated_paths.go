package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// maxSourceLine is the longest added line a code-sink rule still reads.
// Hand-written source almost never passes a few hundred characters; minifier
// output and inlined data are one line of tens of KB. 2000 leaves room for
// long generated-looking string tables in real code while cutting every
// bundle. Secrets and untrusted-diff are not wrapped, so a credential or an
// instruction hidden on a long line is still seen.
const maxSourceLine = 2000

var (
	assetExt     = map[string]bool{".js": true, ".mjs": true, ".cjs": true, ".css": true}
	generatedExt = map[string]bool{".map": true}
	// outputDirs hold build output whatever the file name is.
	outputDirs = map[string]bool{"dist": true, ".next": true}
	// hashSuffix is the content hash bundlers put before the extension
	// (app-88b692ce.js, index-BwK3x9aQ.js, chunk.3f2a9c1b.js).
	hashSuffix = regexp.MustCompile(`[-.]([A-Za-z0-9_]{8,})$`)
)

// isGeneratedPath reports build artifacts: minified files, source maps,
// anything under a dist or .next output dir, public/build assets, and hashed
// bundles under build/. Nobody reviews or edits these by hand, and their
// source of truth is somewhere else in the repo.
func isGeneratedPath(p string) bool {
	lower := strings.ToLower(p)
	ext := path.Ext(lower)
	stem := strings.TrimSuffix(path.Base(lower), ext)
	if generatedExt[ext] || (assetExt[ext] && strings.HasSuffix(stem, ".min")) {
		return true
	}
	if !assetExt[ext] {
		return false
	}
	segs := strings.Split(path.Dir(lower), "/")
	for i, s := range segs {
		switch {
		case outputDirs[s]:
			return true
		case s == "build" && i > 0 && segs[i-1] == "public":
			return true
		case s == "build" && hasContentHash(stem):
			return true
		}
	}
	return false
}

func hasContentHash(stem string) bool {
	m := hashSuffix.FindStringSubmatch(stem)
	return m != nil && strings.ContainsAny(m[1], "0123456789")
}

// withoutLongLines drops hunk lines longer than maxSourceLine so a sink rule
// never reads a minified or inlined blob. Removed lines are kept untouched.
func withoutLongLines(f diff.File) diff.File {
	keep := func(in []diff.Line) []diff.Line {
		out := make([]diff.Line, 0, len(in))
		for _, l := range in {
			if len(l.Text) <= maxSourceLine {
				out = append(out, l)
			}
		}
		return out
	}
	f.Lines, f.Added = keep(f.Lines), keep(f.Added)
	return f
}
