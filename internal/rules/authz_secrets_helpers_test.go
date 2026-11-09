package rules

import (
	"strconv"
	"strings"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// fileOf builds a diff.File from hunk lines prefixed with "+" (added) or " "
// (context). A line "@@ N" jumps the new-file line counter to N (new hunk).
func fileOf(path string, start int, hunk ...string) diff.File {
	f := diff.File{Path: path}
	n := start
	for _, raw := range hunk {
		if strings.HasPrefix(raw, "@@ ") {
			n, _ = strconv.Atoi(strings.TrimPrefix(raw, "@@ "))
			continue
		}
		l := diff.Line{Path: path, Line: n, Text: raw[1:], Kind: diff.Context}
		if raw[0] == '+' {
			l.Kind = diff.Added
			f.Added = append(f.Added, l)
		}
		f.Lines = append(f.Lines, l)
		n++
	}
	return f
}

// handlerFile mirrors the JS test: a context import on line 1, then added lines.
func handlerFile(body ...string) diff.File {
	hunk := []string{" import { db } from '../db';"}
	for _, b := range body {
		hunk = append(hunk, "+"+b)
	}
	return fileOf("src/api/orders.ts", 1, hunk...)
}

func lineNumbers(fs []finding.Finding) []int {
	var out []int
	for _, f := range fs {
		out = append(out, f.Line)
	}
	return out
}

func wantLines(t *testing.T, got []finding.Finding, want ...int) {
	t.Helper()
	g := lineNumbers(got)
	if len(g) != len(want) {
		t.Fatalf("lines = %v, want %v", g, want)
	}
	for i := range want {
		if g[i] != want[i] {
			t.Fatalf("lines = %v, want %v", g, want)
		}
	}
}
