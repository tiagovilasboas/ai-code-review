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
	// scIntegrityValue is the hash itself, without quotes, commas or braces.
	scIntegrityValue = regexp.MustCompile(`sha\d+-[A-Za-z0-9+/=_-]+`)
	// pnpm package key: "  /name@1.2.3:" or "  name@1.2.3:" (indented).
	scPnpmKey = regexp.MustCompile(`^\s{2}'?/?\S+@\S+?'?:\s*$`)
)

// lockfileIntegrityFindings reports an added integrity line whose package
// entry keeps its version (and resolved URL) as unchanged context: the hash
// changed while the version did not, which is how a swapped tarball looks. The
// hash is compared by value: a deleted line with the same hash means no change. A
// version bump changes version, resolved and integrity together and stays
// silent, as does a brand new entry.
func lockfileIntegrityFindings(f diff.File) []finding.Finding {
	var out []finding.Finding
	removed := removedIntegrities(f)
	pnpm := path.Base(f.Path) == "pnpm-lock.yaml"
	for i, l := range f.Lines {
		if l.Kind != diff.Added || !scIntegrityLine.MatchString(l.Text) {
			continue
		}
		if removed[integrityValue(l.Text)] {
			continue // the same hash was already there: a reformat or a moved entry
		}
		if sameVersionEntry(f.Lines, i, pnpm) {
			out = add(out, scInput(f, l.Line, finding.LockfileRewrite, "Lockfile integrity changed without a version change", "CWE-829", finding.High))
		}
	}
	return out
}

func integrityValue(text string) string { return scIntegrityValue.FindString(text) }

// removedIntegrities collects the hash values of deleted integrity lines. An
// added line carrying one of them changed no hash: npm re-emitting an entry
// with a trailing comma, or a package moving to another place in the file.
func removedIntegrities(f diff.File) map[string]bool {
	out := map[string]bool{}
	for _, l := range f.Removed {
		if scIntegrityLine.MatchString(l.Text) {
			if v := integrityValue(l.Text); v != "" {
				out[v] = true
			}
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
