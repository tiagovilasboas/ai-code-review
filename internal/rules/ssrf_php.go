package rules

import (
	"regexp"
	"strings"
)

// PHP outbound HTTP: stream/curl functions, Laravel Http:: and Guzzle-style
// clients. The destination must still trace to request input.
var (
	ssrfPHPFunc = regexp.MustCompile(`\b(?:file_get_contents|fopen|readfile|get_headers|simplexml_load_file|curl_init)\s*\(\s*(.*)$`)
	ssrfPHPCurl = regexp.MustCompile(`\bcurl_setopt\s*\(\s*\$\w+\s*,\s*CURLOPT_URL\s*,\s*(.*)$`)
	ssrfPHPHttp = regexp.MustCompile(`\bHttp::(?:[\w]+\([^()]*\)->)*(?:get|post|put|patch|delete|head)\s*\(\s*(.*)$`)
	// client verbs: ->get($url), ->request('GET', $url), ->send($req).
	ssrfPHPClient = regexp.MustCompile(`\b\w*(?i:client|guzzle|http)\w*->(get|post|put|patch|delete|head|request)\s*\(\s*(.*)$`)
)

// phpLocalPath is a destination that starts as a filesystem path: a path
// constant (WRITEPATH, __DIR__), a path helper (storage_path), realpath (false for a URL) or a relative
// literal with no scheme. Request input after it picks a file, not a host.
var phpLocalPath = regexp.MustCompile(`^(?:realpath\s*\(|(?:[A-Z][A-Z0-9_]*|__DIR__)\s*\.|(?:storage|public|base|app|resource|database)_path\s*\(|dirname\s*\(\s*__FILE__)`)

func phpLocalDest(dest string) bool {
	if phpLocalPath.MatchString(dest) {
		return true
	}
	if dest == "" || (dest[0] != '\'' && dest[0] != '"') {
		return false
	}
	end, from, to := literalBounds(dest, 0)
	lit := dest[from:to]
	return strings.HasPrefix(strings.TrimSpace(dest[end:]), ".") && lit != "" &&
		!strings.Contains(lit, ":") && !strings.HasPrefix(lit, "//")
}

// ssrfPHPDest returns the destination expression of an outbound call on text.
func ssrfPHPDest(text string) (string, bool) {
	for _, re := range []*regexp.Regexp{ssrfPHPFunc, ssrfPHPCurl, ssrfPHPHttp} {
		if m := re.FindStringSubmatch(text); m != nil {
			dest, ok := firstOrEmpty(splitCallArgs(m[1]))
			return dest, ok && !phpLocalDest(dest)
		}
	}
	m := ssrfPHPClient.FindStringSubmatch(text)
	if m == nil {
		return "", false
	}
	args := splitCallArgs(m[2])
	if m[1] == "request" {
		if len(args) < 2 {
			return "", false
		}
		return args[1], true
	}
	return firstOrEmpty(args)
}

func firstOrEmpty(args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	return args[0], true
}
