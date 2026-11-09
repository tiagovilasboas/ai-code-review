package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	scIntegrityLine = regexp.MustCompile(`(?:"integrity"\s*:|^\s*integrity\s|\bintegrity:)\s*"?\{?\s*"?sha`)
	// Lines that identify the package version within one lockfile entry.
	scEntryVersion = regexp.MustCompile(`^\s*(?:"version"\s*:|version\s+"|"resolved"\s*:|resolved\s+")`)
	// pnpm package key: "  /name@1.2.3:" or "  name@1.2.3:" (indented).
	scPnpmKey = regexp.MustCompile(`^\s{2}'?/?\S+@\S+?'?:\s*$`)
)

// lockfileIntegrityFindings reports an added integrity line whose package
// entry keeps its version (and resolved URL) as unchanged context: the hash
// changed while the version did not, which is how a swapped tarball looks. A
// version bump changes version, resolved and integrity together and stays
// silent, as does a brand new entry.
func lockfileIntegrityFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	pnpm := path.Base(f.Path) == "pnpm-lock.yaml"
	for i, l := range f.Lines {
		if l.Kind != diff.Added || !scIntegrityLine.MatchString(l.Text) {
			continue
		}
		if sameVersionEntry(f.Lines, i, pnpm) {
			out = add(out, scInput(f, l.Line, finding.RegistryHost, "Lockfile integrity changed without a version change", "CWE-829", finding.High))
		}
	}
	return out
}

// sameVersionEntry walks back through the contiguous lines of the entry that
// holds lines[idx] and reports whether version evidence is present only as
// context (never added).
func sameVersionEntry(lines []diff.Line, idx int, pnpm bool) bool {
	seenContext := false
	for j := idx - 1; j >= 0 && lines[j].Line == lines[j+1].Line-1; j-- {
		text := lines[j].Text
		isKey := pnpm && scPnpmKey.MatchString(text)
		if isKey || scEntryVersion.MatchString(text) {
			if lines[j].Kind == diff.Added {
				return false
			}
			seenContext = true
			if isKey {
				break
			}
			continue
		}
		if strings.TrimSpace(text) == "" || strings.HasSuffix(strings.TrimSpace(text), "{") || indentOf(text) == 0 {
			break
		}
	}
	return seenContext
}
