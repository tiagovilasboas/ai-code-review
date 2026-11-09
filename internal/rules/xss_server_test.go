package rules

import "testing"

func xssLines(t *testing.T, path string, lines ...string) []int {
	t.Helper()
	var got []int
	for _, f := range XSS(mkFile(path, lines...)) {
		got = append(got, f.Line)
	}
	return got
}

func TestXSSDjangoTemplates(t *testing.T) {
	wantLineNums(t, xssLines(t, "t/a.html", `+<div>{{ post.signature|safe }}</div>`), 1)
	wantLineNums(t, xssLines(t, "t/a.html", `+<p>{{ comment.body|linebreaks|safe }}</p>`), 1)
	wantLineNums(t, xssLines(t, "t/a.html", `+{% autoescape off %}`), 1)
	wantLineNums(t, xssLines(t, "t/a.html", `+<script src="{{ STATIC_URL|safe }}app.js"></script>`))
	wantLineNums(t, xssLines(t, "t/a.html", `+{{ '<br>'|safe }}`))
	wantLineNums(t, xssLines(t, "t/a.html", `+{{ post.title|force_escape|safe }}`))
	wantLineNums(t, xssLines(t, "t/a.html", `+{{ post.body|escape|linebreaks }}`))
	wantLineNums(t, xssLines(t, "t/a.html", `+{% autoescape on %}`))
}

func TestXSSDjangoMarkSafe(t *testing.T) {
	wantLineNums(t, xssLines(t, "app/views.py", `+return render(r, "p.html", {"bio": mark_safe(bio)})`), 1)
	wantLineNums(t, xssLines(t, "app/views.py", `+x = mark_safe(f"<b>{name}</b>")`), 1)
	wantLineNums(t, xssLines(t, "app/views.py", `+x = mark_safe('<b>new</b>')`))
	wantLineNums(t, xssLines(t, "app/views.py", `+x = mark_safe(escape(bio))`))
	wantLineNums(t, xssLines(t, "app/views.py", `+x = mark_safe(format_html("<b>{}</b>", bio))`))
}

func TestXSSExpressResponses(t *testing.T) {
	wantLineNums(t, xssLines(t, "src/r.js", "+res.send(`<h1>Results for ${req.query.q}</h1>`);"), 1)
	wantLineNums(t, xssLines(t, "src/r.js",
		"+const q = req.query.q;",
		"+res.status(200).write(`<p>${q}</p>`);"), 2)
	wantLineNums(t, xssLines(t, "src/r.js", `+res.end('<b>' + req.params.name + '</b>');`), 1)
	wantLineNums(t, xssLines(t, "src/r.js",
		"+const page = `<h1>${req.query.q}</h1>`;",
		"+res.send(page);"), 2)
	wantLineNums(t, xssLines(t, "src/r.js", "+res.json({ h: `<b>${req.query.q}</b>` });"))
	wantLineNums(t, xssLines(t, "src/r.js", "+res.send(`<h1>${escapeHtml(req.query.q)}</h1>`);"))
	wantLineNums(t, xssLines(t, "src/r.js", "+res.send(`<h1>${he.encode(req.query.q)}</h1>`);"))
	wantLineNums(t, xssLines(t, "src/r.js", "+res.send(`<h1>${validator.escape(req.query.q)}</h1>`);"))
	wantLineNums(t, xssLines(t, "src/r.js", "+res.send(`You searched for ${req.query.q}`);"))
	wantLineNums(t, xssLines(t, "src/r.js", "+res.type('text/plain').send(`<b>${req.query.q}</b>`);"))
	wantLineNums(t, xssLines(t, "src/r.js", "+res.send(`<h1>${title}</h1>`);"))
}
