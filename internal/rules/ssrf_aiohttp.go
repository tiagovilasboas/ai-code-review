package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
)

var (
	// pyClientSession marks a hunk that builds an aiohttp (or httpx) client session.
	pyClientSession = regexp.MustCompile(`\b(?:aiohttp\.)?ClientSession\s*\(|\bhttpx\.AsyncClient\s*\(`)
	// pySessionName binds a session variable: `async with ClientSession() as s`, `s = ClientSession()`.
	pySessionName = regexp.MustCompile(`(?:ClientSession|AsyncClient)\s*\([^)]*\)\s+as\s+(\w+)|\b(\w+)\s*=\s*(?:aiohttp\.|httpx\.)?(?:ClientSession|AsyncClient)\s*\(`)
	// ssrfPySession is a request on a variable: s.get(x), self.session.request("GET", x).
	ssrfPySession = regexp.MustCompile(`\b([\w.]+)\.(get|post|put|patch|delete|head|options|request|ws_connect)\s*\(\s*(.*)$`)
)

// pySessionVars returns the names of aiohttp/httpx client sessions in the
// hunk. A bare session.get(x) is a Flask session lookup, so these receivers
// count only when the hunk shows a client session being built.
func pySessionVars(lines []diff.Line) map[string]bool {
	vars := map[string]bool{}
	for _, l := range lines {
		for _, m := range pySessionName.FindAllStringSubmatch(l.Text, -1) {
			vars[m[1]+m[2]] = true
		}
	}
	return vars
}

// ssrfPySessionDest returns the destination of a request on a client session
// variable: session.get(url), session.request("GET", url).
func ssrfPySessionDest(text string, vars map[string]bool) (string, bool) {
	m := ssrfPySession.FindStringSubmatch(text)
	if m == nil || !(vars[m[1]] || strings.Contains(strings.ToLower(m[1]), "session")) {
		return "", false
	}
	args := splitCallArgs(m[3])
	idx := 0
	if m[2] == "request" {
		idx = 1
	}
	if idx >= len(args) {
		return "", false
	}
	return ssrfStripPrefix(args[idx]), true
}
