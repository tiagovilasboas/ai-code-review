package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

// lang is the source language of a changed file, from its extension.
type lang int

const (
	langOther lang = iota
	langJS
	langPython
	langGo
	langJava
	langPHP
)

func langOf(p string) lang {
	switch path.Ext(p) {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return langJS
	case ".py":
		return langPython
	case ".go":
		return langGo
	case ".java":
		return langJava
	case ".php":
		return langPHP
	}
	return langOther
}

var (
	// requestInput is caller-controlled HTTP input in Express/Koa, Flask/Django,
	// net/http/gin/mux/chi, Servlet and PHP/Laravel.
	requestInput = regexp.MustCompile(`\b(?:req|ctx|ctx\.request)\.(?:params|query|body|cookies|headers)\b` +
		`|\brequest\.(?:args|form|values|json|files|cookies|GET|POST|data|get_json|query_params)\b` +
		`|\b(?:request|req)\.getParameter(?:Values)?\s*\(` +
		`|\br\.URL\.Query\(\)|\br\.(?:Form|PostForm)\b|\br\.(?:FormValue|PostFormValue)\s*\(` +
		`|\bmux\.Vars\s*\(|\bchi\.URLParam\s*\(|\bc\.(?:Query|DefaultQuery|Param|Params|PostForm|QueryParam|FormValue)\s*\(` +
		`|\$_(?:GET|POST|REQUEST|COOKIE|FILES)\b|\$request->(?:input|query|get|post)\s*\(`)
	// springParam captures the parameter name bound to request input in Spring and JAX-RS.
	springParam = regexp.MustCompile(`@(?:RequestParam|PathVariable|RequestBody|RequestHeader|PathParam|QueryParam|FormParam|HeaderParam|MatrixParam|CookieParam)\b(?:\s*\([^)]*\))?\s+(?:final\s+)?[\w.<>\[\]]+\s+(\w+)`)

	assignLine = regexp.MustCompile(`^\s*(?:(?:const|let|var|final|val)\s+)?(?:[\w.<>\[\]]+\s+)?((?:\$?[A-Za-z_]\w*\s*,\s*)*\$?[A-Za-z_]\w*)\s*(:=|\+=|\.=|=)\s*([^=>].*)$`)
	// recordLookup is a data-store read keyed by request input: its result is a
	// stored record, not the caller's raw input, so taint stops there.
	recordLookup = regexp.MustCompile(`\b(?:find(?:One|Many|All|Unique|First)?(?:By\w*)?|findById|get\w*By\w+|load\w*By\w+)\s*\(`)
	destructure  = regexp.MustCompile(`^\s*(?:const|let|var)\s*\{([^}]*)\}\s*=\s*(.+)$`)
)

// assignment is one "names op rhs" source line.
type assignment struct {
	names []string
	op    string
	rhs   string
}

// appends reports whether the assignment extends the previous value (+= or .=).
func (a assignment) appends() bool { return a.op == "+=" || a.op == ".=" }

// parseAssignment reads "a = x", "a, b := x", "a += x" and "const { a, b: c } = x".
func parseAssignment(text string) (assignment, bool) {
	if m := destructure.FindStringSubmatch(text); m != nil {
		var names []string
		for _, part := range strings.Split(m[1], ",") {
			if k := strings.Index(part, ":"); k >= 0 {
				part = part[k+1:]
			}
			if k := strings.Index(part, "="); k >= 0 {
				part = part[:k]
			}
			if n := strings.Trim(strings.TrimSpace(part), "."); n != "" {
				names = append(names, n)
			}
		}
		return assignment{names: names, op: "=", rhs: m[2]}, len(names) > 0
	}
	m := assignLine.FindStringSubmatch(text)
	if m == nil {
		return assignment{}, false
	}
	var names []string
	for _, n := range strings.Split(m[1], ",") {
		names = append(names, strings.TrimSpace(n))
	}
	return assignment{names: names, op: m[2], rhs: m[3]}, true
}

// taintedNames returns the variables that hold request input in the visible
// hunk lines: assigned from it, from another tainted variable, or bound to
// it by a Spring annotation. A record loaded by a lookup keyed on input is
// not itself input.
func taintedNames(lines []diff.Line, php bool) map[string]bool {
	tainted := map[string]bool{}
	for _, l := range lines {
		if IsCommentLine(l.Text) {
			continue
		}
		for _, m := range springParam.FindAllStringSubmatch(l.Text, -1) {
			tainted[m[1]] = true
		}
		a, ok := parseAssignment(l.Text)
		if !ok || recordLookup.MatchString(a.rhs) || !reachesRequest(a.rhs, tainted, php) {
			continue
		}
		for _, n := range a.names {
			tainted[n] = true
		}
	}
	return tainted
}

// reachesRequest reports whether text uses request input or a tainted
// variable in code or interpolations (not in plain literal text).
func reachesRequest(text string, tainted map[string]bool, php bool) bool {
	e := scanExpr(text, php)
	hay := e.code + " " + strings.Join(e.interps, " ")
	if requestInput.MatchString(hay) {
		return true
	}
	return mentionsAny(hay, tainted)
}
