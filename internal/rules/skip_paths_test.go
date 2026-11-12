package rules

import (
	"strings"
	"testing"
)

func TestIsGeneratedPath(t *testing.T) {
	cases := map[string]bool{
		"public/build/assets/app-88b692ce.js": true,
		"public/build/assets/app.css":         true,
		"web/dist/index.js":                   true,
		"build/static/main.3f2a9c1b.js":       true,
		"assets/vendor.min.js":                true,
		"assets/site.min.css":                 true,
		"assets/app.js.map":                   true,
		".next/static/chunks/page.js":         true,
		"build/deploy.js":                     false,
		"build/scripts/release.js":            false,
		"src/app.js":                          false,
		"src/bootstrap-datepicker.js":         false,
		"src/admin.min.php":                   false,
		"dist/README.md":                      false,
	}
	for p, want := range cases {
		if got := isGeneratedPath(p); got != want {
			t.Errorf("isGeneratedPath(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestIsTestPathCamelCaseClasses(t *testing.T) {
	cases := map[string]bool{
		"app/Billing/OrderTest.php":             true,
		"src/main/java/OrderServiceTests.java":  true,
		"src/test/java/Foo.java":                true,
		"internal/x/render_test.go":             true,
		"pkg/test_render.py":                    true,
		"app/Latest.php":                        false,
		"app/Contest.php":                       false,
		"src/main/java/test/web/Hello.java":     false,
		"src/main/kotlin/test/routes/Fetch.kt":  false,
		"src/main/java/test/web/HelloTest.java": true,
		"tests/src/main/Foo.java":               true,
	}
	for p, want := range cases {
		if got := isTestPath(p); got != want {
			t.Errorf("isTestPath(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestSkipGeneratedDropsLongLinesOnly(t *testing.T) {
	long := "el.innerHTML = " + strings.Repeat("x", maxSourceLine)
	hit := SkipGenerated(XSS)
	if got := hit(mkFile("public/build/assets/app-88b692ce.js", `+el.innerHTML = a`)); len(got) != 0 {
		t.Fatalf("hashed bundle must be silent, got %+v", got)
	}
	if got := hit(mkFile("src/app.js", `+`+long)); len(got) != 0 {
		t.Fatalf("over-long line must be silent, got %+v", got)
	}
	got := hit(mkFile("src/app.js", `+`+long, `+el.innerHTML = userInput`))
	if len(got) != 1 || got[0].Line != 2 {
		t.Fatalf("normal line after a long one must still fire, got %+v", got)
	}
}

func TestSkipTestsSilencesSinks(t *testing.T) {
	cases := []struct {
		path, line string
		rule       Rule
	}{
		{"internal/render/render_test.go", `+	return template.HTML(body)`, XSS},
		{"tests/Feature/OrderTest.php", `+{!! $request->input('x') !!}`, XSS},
	}
	for _, c := range cases {
		if got := SkipTests(c.rule)(mkFile(c.path, c.line)); len(got) != 0 {
			t.Errorf("%s: want silent, got %+v", c.path, got)
		}
	}
	if got := XSS(mkFile("internal/render/render.go", `+	return template.HTML(body)`)); len(got) != 1 {
		t.Fatalf("non-test file must still fire, got %+v", got)
	}
}

func TestBladePlainTextAndStaticTranslation(t *testing.T) {
	silent := []struct{ path, line string }{
		{"resources/views/email/invoice/invoice_text.blade.php", `+{!! $body !!}`},
		{"resources/views/email/template/text.blade.php", `+{!! $body !!}`},
		{"resources/views/email/x.text.blade.php", `+{!! $body !!}`},
		{"resources/views/email/template/html.blade.php", `+{!! ctrans('texts.thanks') !!}`},
		{"resources/views/page.blade.php", `+{!! __("Welcome") !!}`},
	}
	for _, c := range silent {
		if got := XSS(mkFile(c.path, c.line)); len(got) != 0 {
			t.Errorf("%s %s: want silent, got %+v", c.path, c.line, got)
		}
	}
	fires := []struct{ path, line string }{
		{"resources/views/page.blade.php", `+{!! $request->input('x') !!}`},
		{"resources/views/email/template/html.blade.php", `+{!! $user->bio !!}`},
		{"resources/views/page.blade.php", `+{!! ctrans('texts.' . $key) !!}`},
	}
	for _, c := range fires {
		if got := XSS(mkFile(c.path, c.line)); len(got) != 1 {
			t.Errorf("%s %s: want one finding, got %+v", c.path, c.line, got)
		}
	}
}

func TestSecretsInTestPaths(t *testing.T) {
	silent := []string{
		`+'api_key' => 'sk_test_x',`,
		`+'key' => 'pk_test_51Habcdefghijklmnop',`,
		`+'key' => 'sk_test_51Habcdefghijklmnop',`,
		`+token = "ghp_short"`,
	}
	for _, l := range silent {
		if got := Secrets(mkFile("tests/Unit/PaymentTest.php", l)); len(got) != 0 {
			t.Errorf("%s: want silent, got %+v", l, got)
		}
	}
	if got := Secrets(mkFile("tests/Unit/PaymentTest.php", `+'key' => 'sk_live_Zq81mXv0Lr7Tn4Pw2Kd9',`)); len(got) != 1 {
		t.Fatalf("live-looking key in a test file must fire, got %+v", got)
	}
	if got := Secrets(mkFile("src/pay.php", `+'key' => 'sk_test_x',`)); len(got) != 1 {
		t.Fatalf("outside tests the prefix still fires, got %+v", got)
	}
}
