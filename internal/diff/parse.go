package diff

import (
	"regexp"
	"strconv"
	"strings"
)

var hunkHeader = regexp.MustCompile(`^@@\s+-\d+(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@`)

// Parse reads a unified diff into per-file added/context lines with new-file
// numbers. Only hunk lines on the `+++` side are evidence.
func Parse(text string) []File {
	var files []*File
	var current *File
	newLine := 0
	inHunk := false
	// oldLeft/newLeft count the hunk lines still to read, so "+++ " and "--- "
	// inside a hunk are content, not file headers.
	oldLeft, newLeft := 0, 0

	// JS splits on /\r?\n/; a lone \r is kept.
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		switch {
		case strings.HasPrefix(raw, "diff --git "):
			current, inHunk, oldLeft, newLeft = nil, false, 0, 0
			continue
		case strings.HasPrefix(raw, "+++ ") && oldLeft <= 0 && newLeft <= 0:
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
			oldLeft, newLeft = 0, 0
			if m == nil || current == nil {
				inHunk = false
				continue
			}
			n, err := strconv.Atoi(m[2])
			if err != nil {
				inHunk = false
				continue
			}
			oldLeft, newLeft = hunkCount(m[1]), hunkCount(m[3])
			newLine, inHunk = n, true
			continue
		}

		if !inHunk || current == nil {
			continue
		}

		if strings.HasPrefix(raw, "+") {
			newLeft--
			line := Line{Path: current.Path, Line: newLine, Text: raw[1:], Kind: Added}
			current.Lines = append(current.Lines, line)
			current.Added = append(current.Added, line)
			newLine++
			continue
		}
		if strings.HasPrefix(raw, "-") {
			oldLeft--
			current.Removed = append(current.Removed, Line{Path: current.Path, Line: newLine, Text: raw[1:], Kind: Deleted})
			continue
		}
		if strings.HasPrefix(raw, "\\") {
			continue
		}

		oldLeft--
		newLeft--
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

// hunkCount reads an optional hunk count; an omitted one means a single line.
func hunkCount(s string) int {
	if s == "" {
		return 1
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
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
