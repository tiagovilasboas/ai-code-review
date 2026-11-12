package rules

import "testing"

func TestCORSKtor(t *testing.T) {
	const kt = "src/main/kotlin/com/acme/Application.kt"
	runFwCases(t, Secrets, []fwCase{
		{"any host with credentials", kt, []string{
			`+    install(CORS) {`,
			`+        allowMethod(HttpMethod.Put)`,
			`+        anyHost()`,
			`+        allowCredentials = true`,
			`+    }`}, []int{3}},
		{"allow host star", kt, []string{
			`+    install(CORS) {`,
			`+        allowHost("*")`,
			`+        allowCredentials = true`,
			`+    }`}, []int{2}},
		{"any host without credentials, write methods", kt, []string{
			`+    install(CORS) {`,
			`+        allowMethod(HttpMethod.Delete)`,
			`+        anyHost()`,
			`+    }`}, []int{3}},
		{"any host read only", kt, []string{
			`+    install(CORS) {`,
			`+        allowMethod(HttpMethod.Get)`,
			`+        anyHost()`,
			`+    }`}, nil},
		{"named host with credentials", kt, []string{
			`+    install(CORS) {`,
			`+        allowHost("app.example.test", schemes = listOf("https"))`,
			`+        allowCredentials = true`,
			`+    }`}, nil},
	})
}
