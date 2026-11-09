package rules

import "testing"

func ssrfLines(t *testing.T, path string, lines ...string) []int {
	t.Helper()
	var got []int
	for _, f := range SSRF(mkFile(path, lines...)) {
		got = append(got, f.Line)
	}
	return got
}

func wantLineNums(t *testing.T, got []int, want ...int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("lines = %v, want %v", got, want)
		}
	}
}

func TestSSRFPythonRequestInput(t *testing.T) {
	wantLineNums(t, ssrfLines(t, "app/v.py", `+r = requests.post(request.json["url"], json={}, timeout=5)`), 1)
	wantLineNums(t, ssrfLines(t, "app/v.py",
		`+target = request.form.get("url")`,
		`+r = httpx.get(target)`), 2)
	wantLineNums(t, ssrfLines(t, "app/v.py",
		`+url = request.args["target"]`,
		`+urllib.request.urlopen(url)`), 2)
	wantLineNums(t, ssrfLines(t, "app/v.py",
		`+url = request.args["target"]`,
		`+requests.request("GET", url)`), 2)
}

func TestSSRFPythonSilent(t *testing.T) {
	wantLineNums(t, ssrfLines(t, "app/v.py", `+requests.get("https://api.example.com/health")`))
	wantLineNums(t, ssrfLines(t, "app/v.py",
		`+oid = request.args.get("id")`,
		`+requests.get(f"https://billing.example.com/orders/{oid}")`))
	wantLineNums(t, ssrfLines(t, "app/v.py",
		`+ALLOWED_HOSTS = {"a.example"}`,
		`+url = request.json["url"]`,
		`+if urlparse(url).hostname not in ALLOWED_HOSTS: abort(400)`,
		`+requests.get(url)`))
	wantLineNums(t, ssrfLines(t, "app/v.py", `+requests.get(settings.API_URL)`))
	wantLineNums(t, ssrfLines(t, "app/v.py", `+name = session.get(request.args["k"])`))
}

func TestSSRFGoRequestInput(t *testing.T) {
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+target := r.URL.Query().Get("url")`,
		`+resp, err := http.Get(target)`), 2)
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+dest := r.FormValue("cb")`,
		`+req, _ := http.NewRequest(http.MethodPost, dest, nil)`), 2)
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+var in struct{ URL string }`,
		`+json.NewDecoder(r.Body).Decode(&in)`,
		`+resp, _ := client.Get(in.URL)`), 3)
}

func TestSSRFGoSilent(t *testing.T) {
	wantLineNums(t, ssrfLines(t, "internal/s.go", `+resp, err := http.Get("https://api.example.com/x")`))
	wantLineNums(t, ssrfLines(t, "internal/s.go", `+func fetch(w http.ResponseWriter, r *http.Request) {`))
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+id := r.URL.Query().Get("id")`,
		`+http.Get("https://billing.example.com/o/" + url.PathEscape(id))`))
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+t := r.URL.Query().Get("url")`,
		`+u, _ := url.Parse(t)`,
		`+if !allowedHosts[u.Hostname()] { return }`,
		`+http.Get(u.String())`))
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+t := r.URL.Query().Get("url")`,
		`+if !slices.Contains(allowed, t) { return }`,
		`+http.Get(t)`))
	wantLineNums(t, ssrfLines(t, "internal/s.go",
		`+json.NewDecoder(resp.Body).Decode(&cfg)`,
		`+http.Get(cfg.URL)`))
}
