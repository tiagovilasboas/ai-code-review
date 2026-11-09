// Package diff turns a unified diff into per-file lines with new-file numbers.
// Only lines on the `+++` side are evidence (guardrail: path:line in the diff).
package diff

// LineKind marks whether a hunk line was added or is unchanged context.
type LineKind string

const (
	Added   LineKind = "add"
	Context LineKind = "context"
)

// Line is one hunk line on the new side of the diff.
type Line struct {
	Path string
	Line int
	Text string
	Kind LineKind
}

// File holds every new-side hunk line of one changed file, plus the added subset.
type File struct {
	Path  string
	Lines []Line
	Added []Line
}
