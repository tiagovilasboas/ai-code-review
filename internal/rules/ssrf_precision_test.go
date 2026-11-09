package rules

import "testing"

func TestSSRFFlagsSteerableDestinations(t *testing.T) {
	calls := []string{
		`+await axios({ method: 'get', url: req.body.u });`,
		`+await axios.request({ url: req.query.u });`,
		`+http.get('http://' + req.query.host + '/x', cb);`,
		"+https.get(`https://${req.query.h}/x`, cb);",
		`+http.request(req.body.url, opts, cb);`,
		`+await got.post(req.body.cb, { json: {} });`,
		`+await fetch(req.query.u as string);`,
		"+await fetch(`https://api.${req.query.d}.com/x`);",
	}
	for _, c := range calls {
		got := SSRF(mkFile("src/server/x.js", c))
		if len(got) != 1 {
			t.Errorf("%q: want 1 finding, got %d", c, len(got))
		}
	}
}

func TestSSRFSilentOnFixedDestinations(t *testing.T) {
	calls := []string{
		"+await fetch(`https://api.example.invalid/v1/${encodeURIComponent(req.params.id)}`);",
		`+await fetch('https://api.example.invalid/u/' + Number(req.params.id));`,
		`+await fetch(process.env.HOOK_URL, { method: 'POST' });`,
		"+await fetch(`${process.env.BASE}/x`);",
		`+await fetch(BASE_URL + '/x');`,
		`+await axios({ method: 'get', url: 'https://x.example.invalid' });`,
		`+await axios({ method: 'get' });`,
	}
	for _, c := range calls {
		if got := SSRF(mkFile("src/server/x.js", c)); len(got) != 0 {
			t.Errorf("%q: want none, got %+v", c, got)
		}
	}
}

func TestSSRFAllowlistEarlierInHunkSilencesLaterCall(t *testing.T) {
	f := mkFile("src/server/x.js",
		`+const u = new URL(req.query.u);`,
		`+if (!ALLOWED.has(u.hostname)) return res.sendStatus(403);`,
		`+await fetch(u);`,
	)
	if got := SSRF(f); len(got) != 0 {
		t.Fatalf("want none, got %+v", got)
	}
	g := mkFile("src/server/x.js", `+await fetch(u);`, `+if (!ALLOWED.has(u.hostname)) return;`)
	if got := SSRF(g); len(got) != 1 {
		t.Fatalf("check after the call must not count, got %+v", got)
	}
}
