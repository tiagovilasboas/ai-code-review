package rules

import (
	"regexp"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var (
	// sqlSink is a call that sends its string argument to a database driver or ORM.
	sqlSink = regexp.MustCompile(`(?:^|[^\w$])\$?(?:execute|executemany|executescript|query|raw|exec|` +
		`Exec|ExecContext|Query|QueryContext|QueryRow|QueryRowContext|Raw|` +
		`createQuery|createNativeQuery|createSQLQuery|executeQuery|executeUpdate|executeLargeUpdate|` +
		`prepareStatement|prepare|queryForObject|queryForList|queryForMap|queryRawUnsafe|executeRawUnsafe|` +
		`mysqli_query|mysql_query|pg_query|sqlite_query)\s*\(`)
	// sqlKeyword marks literal text as SQL.
	sqlKeyword = regexp.MustCompile(`(?i)\bselect\b[\s\S]*\bfrom\b|\binsert\s+into\b|\bupdate\b[\s\S]*\bset\b|` +
		`\bdelete\s+from\b|\bwhere\b|\border\s+by\b|\bgroup\s+by\b|\bvalues\s*\(`)
	// sqlEscaped is a visible escape or numeric cast: precision first, stay silent.
	sqlEscaped = regexp.MustCompile(`\b(?:mysqli_real_escape_string|mysql_real_escape_string|pg_escape_string|` +
		`pg_escape_literal|escape|escapeId|escapeLiteral|quote|intval|parseInt|Number|int|Atoi|parseLong)\s*\(`)
)

// SQLInjection flags SQL text built from non-literal values by concatenation,
// interpolation or formatting and passed to a query call (CWE-89). It is HIGH
// when request input visibly reaches the query, MEDIUM otherwise.
// Parameterized queries (placeholders plus an argument list) never fire.
func SQLInjection(f diff.File) []finding.Finding {
	if langOf(f.Path) == langOther {
		return nil
	}
	php := langOf(f.Path) == langPHP
	tainted := taintedNames(f.Lines, php)
	safe := allowlistedValues(f.Lines)
	built := builtSQLVars(f.Lines, php, tainted, safe)
	var out []finding.Finding
	for i, l := range f.Lines {
		if l.Kind != diff.Added || IsCommentLine(l.Text) {
			continue
		}
		loc := sqlSink.FindStringIndex(l.Text)
		if loc == nil {
			continue
		}
		args := callArgs(callFrom(f.Lines, i, 0), loc[1]-1)
		dynamic, fromRequest := false, false
		for _, a := range args {
			d, t := sqlArgBuilt(a, php, built, safe)
			dynamic = dynamic || d
			fromRequest = fromRequest || t || (d && reachesRequest(a, tainted, php))
		}
		if !dynamic {
			continue
		}
		sev := finding.Medium
		if fromRequest {
			sev = finding.High
		}
		out = add(out, finding.Input{
			Title:    "SQL built from non-literal input",
			Class:    finding.SQLInjection,
			Path:     f.Path,
			Line:     l.Line,
			CWE:      "CWE-89",
			OWASP:    finding.LabelsFor(finding.SQLInjection),
			Severity: sev,
		})
	}
	return out
}

// sqlArgBuilt reports whether a query argument is SQL built from non-literal
// values, and whether a variable it uses was built from request input.
func sqlArgBuilt(a string, php bool, built, safe map[string]bool) (dynamic, fromRequest bool) {
	if sqlEscaped.MatchString(a) {
		return false, false
	}
	e := scanExpr(a, php)
	if len(dropAllowlisted(e.dynamicOperands(php), safe)) > 0 && sqlKeyword.MatchString(e.literal) {
		dynamic = true
	}
	for name, tainted := range built {
		if mentions(e.code, name) {
			dynamic = true
			fromRequest = fromRequest || tainted
		}
	}
	return dynamic, fromRequest
}

// builtSQLVars maps each variable holding SQL built from non-literal values
// to whether request input reaches it. A plain reassignment clears it.
func builtSQLVars(lines []diff.Line, php bool, tainted, safe map[string]bool) map[string]bool {
	sqlVars := map[string]bool{}
	built := map[string]bool{}
	for _, l := range lines {
		if IsCommentLine(l.Text) {
			continue
		}
		a, ok := parseAssignment(l.Text)
		if !ok || len(a.names) != 1 {
			continue
		}
		name := a.names[0]
		e := scanExpr(a.rhs, php)
		isSQL := sqlKeyword.MatchString(e.literal) || mentionsAny(e.code, sqlVars) || (a.appends() && sqlVars[name])
		ops := dropAllowlisted(e.dynamicOperands(php), safe)
		if a.appends() {
			ops = dropAllowlisted(e.nonLiteralOperands(), safe)
		}
		usesBuilt := mentionsAny(e.code, built)
		dynamic := (len(ops) > 0 && !sqlEscaped.MatchString(a.rhs)) || usesBuilt || (a.appends() && hasKey(built, name))
		if isSQL {
			sqlVars[name] = true
		}
		if !isSQL || !dynamic {
			if !a.appends() {
				delete(built, name)
			}
			continue
		}
		built[name] = built[name] || reachesRequest(a.rhs, tainted, php) || builtFromRequest(e.code, built)
	}
	return built
}

func hasKey(m map[string]bool, k string) bool {
	_, ok := m[k]
	return ok
}

func builtFromRequest(code string, built map[string]bool) bool {
	for name, tainted := range built {
		if tainted && mentions(code, name) {
			return true
		}
	}
	return false
}
