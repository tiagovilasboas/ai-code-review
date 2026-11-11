package diff

import (
	"regexp"
	"strconv"
	"strings"
)

var hunkHeader = regexp.MustCompile(`^@@\s+-\d+(?:,\d+)?\s+\+(\d+)(?:,\d+)?\s+@@`)

// Parse reads a unified diff into per-file added/context lines with new-file
// numbers. It is a port of src/parse-diff.js; only hunk lines on the `+++`
// side are evidence.
func Parse(text string) []File {
	var files []*File
	var current *File
	newLine := 0
	inHunk := false

	// JS splits on /\r?\n/; a lone \r is kept.
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		switch {
		case strings.HasPrefix(raw, "diff --git "):
			current, inHunk = nil, false
			continue
		case strings.HasPrefix(raw, "+++ "):
			spec := strings.TrimSpace(raw[4:])
			inHunk = false
			if spec == "/dev/null" {
				current = nil
				continue
			}
			current = &File{Path: strings.TrimPrefix(spec, "b/")}
			files = append(files, current)
			continue
		case strings.HasPrefix(raw, "@@ "):
			m := hunkHeader.FindStringSubmatch(raw)
			if m == nil || current == nil {
				inHunk = false
				continue
			}
			n, err := strconv.Atoi(m[1])
			if err != nil {
				inHunk = false
				continue
			}
			newLine, inHunk = n, true
			continue
		}

		if !inHunk || current == nil {
			continue
		}

		if strings.HasPrefix(raw, "+") {
			line := Line{Path: current.Path, Line: newLine, Text: raw[1:], Kind: Added}
			current.Lines = append(current.Lines, line)
			current.Added = append(current.Added, line)
			newLine++
			continue
		}
		if strings.HasPrefix(raw, "-") {
			current.Removed = append(current.Removed, Line{Path: current.Path, Line: newLine, Text: raw[1:], Kind: Deleted})
			continue
		}
		if strings.HasPrefix(raw, "\\") {
			continue
		}

		current.Lines = append(current.Lines, Line{
			Path: current.Path,
			Line: newLine,
			Text: strings.TrimPrefix(raw, " "),
			Kind: Context,
		})
		newLine++
	}

	out := make([]File, 0, len(files))
	for _, f := range files {
		anchorRemoved(f)
		out = append(out, *f)
	}
	return out
}

// anchorRemoved moves each deleted line to the nearest new-side line at or
// before it, so the anchor always lies inside a hunk (a deletion at the end of
// a hunk would otherwise point one line past it).
func anchorRemoved(f *File) {
	seen := make(map[int]bool, len(f.Lines))
	for _, l := range f.Lines {
		seen[l.Line] = true
	}
	for i, r := range f.Removed {
		for n := r.Line; n >= 1; n-- {
			if seen[n] {
				f.Removed[i].Line = n
				break
			}
		}
	}
}
