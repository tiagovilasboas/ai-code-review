package rules

import "testing"

func TestSSRFFrameworks(t *testing.T) {
	const kt = "src/main/kotlin/com/acme/Proxy.kt"
	const php = "app/Controllers/Preview.php"
	const py = "app/handlers.py"
	const gof = "internal/proxy/forward.go"
	runFwCases(t, SSRF, []fwCase{
		// Ktor
		{"ktor query url", kt, []string{
			`+    val target = call.request.queryParameters["url"] ?: return@get`,
			`+    val body: String = httpClient.get(target).body()`}, []int{2}},
		{"ktor parameters url post", kt, []string{
			`+    val target: String = call.parameters["url"]!!`,
			`+    client.post(target) { setBody("x") }`}, []int{2}},
		{"ktor receiveParameters", kt, []string{
			`+    val params = call.receiveParameters()`,
			`+    val r = httpClient.request(params["url"]!!)`}, []int{2}},
		{"ktor named url argument", kt, []string{
			`+    val u = call.parameters["u"]!!`,
			`+    httpClient.get(urlString = u)`}, []int{2}},
		{"ktor template host", kt, []string{
			`+    val h = call.parameters["host"]!!`,
			`+    httpClient.get("https://$h/status")`}, []int{2}},
		{"ktor fixed host path segment", kt, []string{
			`+    val id = call.parameters["id"]!!`,
			`+    httpClient.get("https://api.partner.example.test/items/$id")`}, nil},
		{"ktor host allowlist", kt, []string{
			`+    val target = call.parameters["url"]!!`,
			`+    if (Url(target).host !in allowedHosts) return@get`,
			`+    httpClient.get(target)`}, nil},
		{"ktor constant url", kt, []string{
			`+    httpClient.get("https://api.partner.example.test/health")`}, nil},
		// CodeIgniter 4
		{"ci4 curl getGet", php, []string{
			`+    curl_setopt($ch, CURLOPT_URL, $this->request->getGet('url'));`}, []int{1}},
		{"ci4 curl getVar via variable", php, []string{
			`+    $u = $this->request->getVar('url');`,
			`+    $r = file_get_contents($u);`}, []int{2}},
		{"ci4 service request", php, []string{
			`+    $u = service('request')->getPost('url');`,
			`+    curl_setopt($ch, CURLOPT_URL, $u);`}, []int{2}},
		{"ci4 curl constant url", php, []string{
			`+    curl_setopt($ch, CURLOPT_URL, 'https://api.partner.example.test/v1');`}, nil},
		{"ci4 readfile path constant is a file read", php, []string{
			`+    readfile(WRITEPATH . 'uploads/' . $this->request->getGet('file'));`}, nil},
		{"php relative path literal is a file read", php, []string{
			`+    $d = file_get_contents('cache/' . $_GET['k']);`}, nil},
		{"php scheme literal prefix still fires", php, []string{
			`+    $d = file_get_contents('https://' . $_GET['host'] . '/x');`}, []int{1}},
		// aiohttp
		{"aiohttp session get", py, []string{
			`+    url = request.query["url"]`,
			`+    async with aiohttp.ClientSession() as session:`,
			`+        async with session.get(url, timeout=5) as resp:`}, []int{3}},
		{"aiohttp session request", py, []string{
			`+    data = await request.json()`,
			`+    async with aiohttp.ClientSession() as s:`,
			`+        resp = await s.request("GET", data["url"])`}, []int{3}},
		{"aiohttp match_info", py, []string{
			`+    host = request.match_info["host"]`,
			`+    async with aiohttp.ClientSession() as session:`,
			`+        await session.get(f"http://{host}/x")`}, []int{3}},
		{"aiohttp host allowlist", py, []string{
			`+    url = request.query["url"]`,
			`+    if urlparse(url).hostname not in ALLOWED_HOSTS: raise web.HTTPForbidden()`,
			`+    async with aiohttp.ClientSession() as session:`,
			`+        await session.get(url)`}, nil},
		{"flask session lookup stays silent", py, []string{
			`+    name = session.get(request.args["k"])`}, nil},
		// Go: parsed host compared with a constant
		{"go hostname constant compare", gof, []string{
			`+	target := r.URL.Query().Get("t")`,
			`+	u, err := url.Parse(target)`,
			`+	if err != nil || u.Scheme != "https" || u.Hostname() != "hooks.partner.example.test" {`,
			`+		return`,
			`+	}`,
			`+	resp, err := http.Get(u.String())`}, nil},
		{"go host constant compare", gof, []string{
			`+	target := r.URL.Query().Get("t")`,
			`+	u, _ := url.Parse(target)`,
			`+	if u.Host != "hooks.partner.example.test" {`,
			`+		return`,
			`+	}`,
			`+	resp, err := http.Get(u.String())`}, nil},
		{"go hostname map lookup", gof, []string{
			`+	target := r.URL.Query().Get("t")`,
			`+	u, _ := url.Parse(target)`,
			`+	if _, ok := partners[u.Hostname()]; !ok {`,
			`+		return`,
			`+	}`,
			`+	resp, err := http.Get(u.String())`}, nil},
		{"go scheme and empty host only", gof, []string{
			`+	target := r.URL.Query().Get("t")`,
			`+	u, err := url.Parse(target)`,
			`+	if err != nil || u.Scheme != "https" || u.Host == "" {`,
			`+		return`,
			`+	}`,
			`+	resp, err := http.Get(u.String())`}, []int{6}},
		{"go host compare after the request does not count", gof, []string{
			`+	target := r.URL.Query().Get("t")`,
			`+	u, _ := url.Parse(target)`,
			`+	resp, err := http.Get(u.String())`,
			`+	if u.Host != "hooks.partner.example.test" {`,
			`+		return`,
			`+	}`}, []int{3}},
	})
}
