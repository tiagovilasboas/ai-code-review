package rules

import "testing"

func TestXSSPrecisionSilent(t *testing.T) {
	cases := map[string][]string{
		"a.js":                    {`+el.innerHTML = '';`},
		"b.js":                    {"+el.innerHTML = `<p>static</p>`;"},
		"c.js":                    {`+el.innerHTML = DOMPurify.sanitize(raw);`},
		"d.js":                    {`+el.innerHTML = sanitizeHtml(raw, { a: 1 });`},
		"e.tsx":                   {`+<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(x) }} />`},
		"f.tsx":                   {"+const clean = DOMPurify.sanitize(x);", "+<div dangerouslySetInnerHTML={{ __html: clean }} />"},
		"g.js":                    {`+$(el).html();`},
		"h.js":                    {`+$(el).html('<b>x</b>');`},
		"i.js":                    {`+root.insertAdjacentHTML('beforeend', '<hr>');`},
		"j.tsx":                   {"+<a href={`/o/${o.id}`}>x</a>"},
		"k.tsx":                   {"+const ok = isSafeUrl(p.website);", "+<a href={p.website}>x</a>"},
		"test/helpers/mount.js":   {"+el.innerHTML = `<i>${x}</i>`;"},
		"src/a/b.test.tsx":        {`+el.innerHTML = x;`},
		"src/__tests__/render.js": {`+el.innerHTML = x;`},
	}
	for path, lines := range cases {
		if got := XSS(mkFile(path, lines...)); len(got) != 0 {
			t.Errorf("%s %q: want none, got %+v", path, lines, got)
		}
	}
}

func TestXSSPrecisionFlags(t *testing.T) {
	cases := map[string][]string{
		"a.js":  {`+el.innerHTML = '<b>' + name + '</b>';`},
		"b.js":  {`+el.innerHTML = a + DOMPurify.sanitize(b);`},
		"c.js":  {`+$('#x').html(data.message);`},
		"d.js":  {`+$el.html('<b>' + t + '</b>');`},
		"e.tsx": {"+const clean = DOMPurify.sanitize(a);", "+<div dangerouslySetInnerHTML={{ __html: other }} />"},
		"f.tsx": {`+<a href={profile.website}>x</a>`},
		"g.vue": {`+<a :href="c.homepageUrl">x</a>`},
		"h.js":  {"+root.insertAdjacentHTML('beforeend', `<p>${m}</p>`);"},
	}
	for path, lines := range cases {
		if got := XSS(mkFile(path, lines...)); len(got) == 0 {
			t.Errorf("%s %q: want a finding", path, lines)
		}
	}
}
