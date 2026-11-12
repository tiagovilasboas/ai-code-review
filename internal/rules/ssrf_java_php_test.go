package rules

import "testing"

func TestSSRFJavaPHPGoPython(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       []int
	}{
		{"java RestTemplate with request param", "src/Proxy.java", []string{
			`+public String fetch(@RequestParam String url) {`,
			`+  return restTemplate.getForObject(url, String.class);`,
			`+}`}, []int{2}},
		{"java new URL openStream on path variable", "src/Proxy.java", []string{
			`+public void get(@PathVariable String target) throws IOException {`,
			`+  InputStream in = new URL(target).openStream();`,
			`+}`}, []int{2}},
		{"java WebClient uri from body dto", "src/Hook.java", []string{
			`+public Mono<String> send(@RequestBody HookDto body) {`,
			`+  return webClient.post().uri(body.getCallbackUrl()).retrieve().bodyToMono(String.class);`,
			`+}`}, []int{2}},
		{"java HttpRequest URI.create from getParameter", "src/S.java", []string{
			`+String u = request.getParameter("url");`,
			`+HttpRequest r = HttpRequest.newBuilder(URI.create(u)).build();`}, []int{2}},
		{"java Apache HttpGet", "src/S.java", []string{
			`+public void f(@RequestParam String link) {`,
			`+  HttpGet get = new HttpGet(link);`,
			`+}`}, []int{2}},
		{"java fixed host with id", "src/Proxy.java", []string{
			`+public String f(@PathVariable String id) {`,
			`+  return restTemplate.getForObject("https://billing.example.com/orders/" + id, String.class);`,
			`+}`}, nil},
		{"java literal url", "src/Proxy.java", []string{
			`+return restTemplate.getForObject("https://api.example.com/health", String.class);`}, nil},
		{"java allowlist first", "src/Proxy.java", []string{
			`+public String f(@RequestParam String url) {`,
			`+  if (!allowedHosts.contains(URI.create(url).getHost())) throw new IllegalArgumentException();`,
			`+  return restTemplate.getForObject(url, String.class);`,
			`+}`}, nil},
		{"java configured base", "src/Proxy.java", []string{
			`+public String f(@RequestParam String q) {`,
			`+  return restTemplate.getForObject(env.getProperty("svc.url"), String.class);`,
			`+}`}, nil},
		{"java uri builder lambda", "src/Proxy.java", []string{
			`+public Mono<String> f(@PathVariable String id) {`,
			`+  return webClient.get().uri(u -> u.path("/orders/{id}").build(id)).retrieve().bodyToMono(String.class);`,
			`+}`}, nil},
		{"java test file", "src/test/ProxyTest.java", []string{
			`+public void f(@RequestParam String url) {`,
			`+  restTemplate.getForObject(url, String.class);`,
			`+}`}, nil},
		{"php file_get_contents on request input", "app/Http/Controllers/P.php", []string{
			`+$url = $request->input('url');`,
			`+return file_get_contents($url);`}, []int{2}},
		{"php curl_setopt from GET", "api/p.php", []string{
			`+curl_setopt($ch, CURLOPT_URL, $_GET['target']);`}, []int{1}},
		{"php laravel Http facade", "app/Http/Controllers/P.php", []string{
			`+$target = $request->query('u');`,
			`+$res = Http::withToken($t)->get($target);`}, []int{2}},
		{"php guzzle request", "app/Services/G.php", []string{
			`+$res = $this->client->request('GET', $request->input('callback'));`}, []int{1}},
		{"php fixed host with id", "app/Http/Controllers/P.php", []string{
			`+$id = $request->input('id');`,
			`+$res = Http::get("https://billing.example.com/orders/$id");`}, nil},
		{"php config url", "app/Http/Controllers/P.php", []string{
			`+$res = Http::get(config('services.billing.url'));`}, nil},
		{"php allowlist first", "app/Http/Controllers/P.php", []string{
			`+$url = $request->input('url');`,
			`+if (!in_array(parse_url($url, PHP_URL_HOST), $allowedHosts)) abort(422);`,
			`+return file_get_contents($url);`}, nil},
		{"php non http request->get", "app/Http/Controllers/P.php", []string{
			`+$v = $request->get('url');`}, nil},
		{"go resty chain", "h.go", []string{
			`+u := r.URL.Query().Get("url")`,
			`+resp, err := client.R().SetHeader("A", "b").Get(u)`}, []int{2}},
		{"go resty fixed", "h.go", []string{
			`+resp, err := client.R().Get("https://api.example.com/v1/x")`}, nil},
		{"python urlretrieve", "app/v.py", []string{
			`+u = request.args["url"]`,
			`+urllib.request.urlretrieve(u, "/tmp/x")`}, []int{2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantLineNums(t, ssrfLines(t, c.path, c.body...), c.want...)
		})
	}
}
