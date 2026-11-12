package rules

import "testing"

func TestSSRFParseThenGuard(t *testing.T) {
	const java = "src/main/java/acme/web/Proxy.java"
	const php = "app/Controllers/Mirror.php"
	runFwCases(t, SSRF, []fwCase{
		{"java uri parsed then host allowlist", java, []string{
			`     public String get(@QueryParam("dest") String dest) throws Exception {`,
			`+        URI u = URI.create(dest);`,
			`+        if (!HOSTS_ALLOWED.contains(u.getHost())) {`,
			`+            throw new ForbiddenException();`,
			`+        }`,
			`+        http.send(HttpRequest.newBuilder(u).build(), BodyHandlers.ofString());`}, nil},
		{"java uri parsed then host set lookup", java, []string{
			`     public String get(@QueryParam("dest") String dest) throws Exception {`,
			`+        URI u = URI.create(dest);`,
			`+        if (!trusted.contains(u.getHost())) { throw new ForbiddenException(); }`,
			`+        http.send(HttpRequest.newBuilder(u).build(), BodyHandlers.ofString());`}, nil},
		{"java uri parsed and never checked", java, []string{
			` public String get(@QueryParam("dest") String dest) throws Exception {`,
			`+        URI u = URI.create(dest);`,
			`+        log.info("fetching {}", u);`}, []int{2}},
		{"php host in literal list", php, []string{
			`     $target = $this->request->getGet('u');`,
			`+    if (!in_array(parse_url($target, PHP_URL_HOST), ['a.example.test'], true)) {`,
			`+        return $this->response->setStatusCode(403);`,
			`+    }`,
			`+    $out = file_get_contents($target);`}, nil},
		{"php host var in list", php, []string{
			`     $target = $this->request->getGet('u');`,
			`+    $targetHost = parse_url($target, PHP_URL_HOST);`,
			`+    if (!in_array($targetHost, $hosts, true)) { return; }`,
			`+    $out = file_get_contents($target);`}, nil},
		{"php in_array on another value is no guard", php, []string{
			`     $target = $this->request->getGet('u');`,
			`+    if (!in_array($mode, ['a', 'b'], true)) { return; }`,
			`+    $out = file_get_contents($target);`}, []int{3}},
		{"php realpath result is a local file", php, []string{
			`     $name = basename((string) $this->request->getGet('f'));`,
			`+    $full = realpath(WRITEPATH . 'exports/' . $name);`,
			`+    readfile($full);`}, nil},
		{"php request url still fetched", php, []string{
			`     $target = $this->request->getGet('u');`,
			`+    readfile($target);`}, []int{2}},
	})
}

func TestPathTraversalGoRootClean(t *testing.T) {
	const gof = "internal/static/serve.go"
	runFwCases(t, PathTraversal, []fwCase{
		{"clean against root before join", gof, []string{
			` func (s *Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {`,
			`+	p := filepath.Join(s.dir, filepath.Clean("/"+r.URL.Query().Get("f")))`,
			`+	b, _ := os.ReadFile(p)`}, nil},
		{"path clean with separator", gof, []string{
			` func (s *Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {`,
			`+	p := filepath.Join(s.dir, path.Clean(string(filepath.Separator)+r.PathValue("f")))`,
			`+	b, _ := os.ReadFile(p)`}, nil},
		{"plain clean keeps dot-dot", gof, []string{
			` func (s *Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {`,
			`+	p := filepath.Join(s.dir, filepath.Clean(r.URL.Query().Get("f")))`,
			`+	b, _ := os.ReadFile(p)`}, []int{3}},
	})
}

func TestXSSTestFileUnderProductionRoot(t *testing.T) {
	cases := map[string]bool{
		"src/main/java/acme/test/web/Hello.java": false,
		"src/main/java/acme/web/Hello.java":      false,
		"src/test/java/acme/web/HelloTest.java":  true,
		"tests/src/main/java/acme/Hello.java":    true,
		"src/main/js/widgets/card.stories.tsx":   true,
		"web/src/__tests__/card.tsx":             true,
		"src/main/resources/fixtures/page.html":  false,
	}
	for p, want := range cases {
		if got := xssTestFile(p); got != want {
			t.Errorf("xssTestFile(%q) = %v, want %v", p, got, want)
		}
	}
	runFwCases(t, XSS, []fwCase{
		{"vertx lambda in a package named test", "src/main/java/acme/test/web/Greet.java", []string{
			`+        router.get("/greet").handler(rc -> {`,
			`+            String who = rc.request().getParam("who");`,
			`+            rc.response().end("<p>Hi " + who + "</p>");`,
			`+        });`}, []int{3}},
	})
}
