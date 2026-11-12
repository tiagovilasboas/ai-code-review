package rules

import "regexp"

// Java outbound HTTP: URL/URI constructors, Apache HttpClient requests,
// WebClient/RestClient .uri(...) and RestTemplate verbs. The destination is
// the first argument; ssrfLangLines still requires it to trace to request
// input (@RequestParam, @PathVariable, @RequestBody, getParameter).
var ssrfJavaCall = regexp.MustCompile(`(?:\bnew\s+(?:URL|URI|HttpGet|HttpPost|HttpPut|HttpPatch|HttpDelete|HttpHead)|\bURI\.create|\.uri|\b\w*(?i:resttemplate|restclient|webclient)\w*\.(?:getForObject|getForEntity|postForObject|postForEntity|postForLocation|exchange|execute|put|delete|headForHeaders))\s*\(\s*()(.*)$`)
