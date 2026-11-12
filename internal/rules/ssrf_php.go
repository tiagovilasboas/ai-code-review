package rules

import "regexp"

// PHP outbound HTTP: stream/curl functions, Laravel Http:: and Guzzle-style
// clients. The destination must still trace to request input.
var (
	ssrfPHPFunc = regexp.MustCompile(`\b(?:file_get_contents|fopen|readfile|get_headers|simplexml_load_file|curl_init)\s*\(\s*(.*)$`)
	ssrfPHPCurl = regexp.MustCompile(`\bcurl_setopt\s*\(\s*\$\w+\s*,\s*CURLOPT_URL\s*,\s*(.*)$`)
	ssrfPHPHttp = regexp.MustCompile(`\bHttp::(?:[\w]+\([^()]*\)->)*(?:get|post|put|patch|delete|head)\s*\(\s*(.*)$`)
	// client verbs: ->get($url), ->request('GET', $url), ->send($req).
	ssrfPHPClient = regexp.MustCompile(`\b\w*(?i:client|guzzle|http)\w*->(get|post|put|patch|delete|head|request)\s*\(\s*(.*)$`)
)

// ssrfPHPDest returns the destination expression of an outbound call on text.
func ssrfPHPDest(text string) (string, bool) {
	for _, re := range []*regexp.Regexp{ssrfPHPFunc, ssrfPHPCurl, ssrfPHPHttp} {
		if m := re.FindStringSubmatch(text); m != nil {
			args := splitCallArgs(m[1])
			return firstOrEmpty(args)
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
