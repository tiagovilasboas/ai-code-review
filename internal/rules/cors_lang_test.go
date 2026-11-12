package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestCORSLanguages(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		// Python
		{"fastapi wildcard with credentials", "app/main.py", []string{
			`+app.add_middleware(CORSMiddleware, allow_origins=["*"],`,
			`+    allow_credentials=True, allow_methods=["*"])`}, 1},
		{"fastapi wildcard read only", "app/main.py", []string{
			`+app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["GET"])`}, 0},
		{"fastapi explicit origins", "app/main.py", []string{
			`+app.add_middleware(CORSMiddleware, allow_origins=["https://app.example.com"], allow_credentials=True)`}, 0},
		{"flask-cors origins star", "app/main.py", []string{`+CORS(app, origins="*", methods=["POST"])`}, 1},
		{"flask-cors resources dict", "app/main.py", []string{`+CORS(app, resources={r"/api/*": {"origins": "*"}}, methods=["POST"])`}, 1},
		{"flask-cors credentials default origins", "app/main.py", []string{`+CORS(app, supports_credentials=True)`}, 1},
		{"flask-cors fixed origin", "app/main.py", []string{`+CORS(app, origins=["https://a.example.com"], supports_credentials=True)`}, 0},
		{"django allow all", "app/settings.py", []string{`+CORS_ALLOW_ALL_ORIGINS = True`, `+CORS_ALLOW_CREDENTIALS = True`}, 1},
		{"django legacy allow all", "app/settings.py", []string{`+CORS_ORIGIN_ALLOW_ALL = True`}, 1},
		{"django allowlist", "app/settings.py", []string{`+CORS_ALLOWED_ORIGINS = ["https://a.example.com"]`, `+CORS_ALLOW_CREDENTIALS = True`}, 0},
		{"django reflect META", "app/views.py", []string{
			`+origin = request.META.get("HTTP_ORIGIN")`,
			`+response["Access-Control-Allow-Origin"] = origin`,
			`+response["Access-Control-Allow-Credentials"] = "true"`}, 1},
		{"flask reflect headers", "app/views.py", []string{
			`+resp.headers["Access-Control-Allow-Origin"] = request.headers.get("Origin")`,
			`+resp.headers["Access-Control-Allow-Credentials"] = "true"`}, 1},
		{"flask reflect after allowlist", "app/views.py", []string{
			`+origin = request.headers.get("Origin")`,
			`+if origin in ALLOWED_ORIGINS:`,
			`+    resp.headers["Access-Control-Allow-Origin"] = origin`}, 0},
		// Go
		{"gin cors AllowAllOrigins", "cmd/s.go", []string{
			`+r.Use(cors.New(cors.Config{AllowAllOrigins: true, AllowCredentials: true}))`}, 1},
		{"gin cors AllowOrigins star", "cmd/s.go", []string{
			`+r.Use(cors.New(cors.Config{AllowOrigins: []string{"*"}, AllowMethods: []string{"POST"}}))`}, 1},
		{"rs/cors AllowedOrigins star", "cmd/s.go", []string{
			`+h := cors.New(cors.Options{AllowedOrigins: []string{"*"}, AllowCredentials: true})`}, 1},
		{"rs/cors AllowOriginFunc true", "cmd/s.go", []string{
			`+cors.Options{AllowOriginFunc: func(o string) bool { return true }, AllowedMethods: []string{"PUT"}}`}, 1},
		{"gin cors explicit origin", "cmd/s.go", []string{
			`+r.Use(cors.New(cors.Config{AllowOrigins: []string{"https://a.example.com"}, AllowCredentials: true}))`}, 0},
		{"go gin header echo", "cmd/s.go", []string{
			`+c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))`}, 1},
		// Java
		{"spring addAllowedOrigin star", "A.java", []string{
			`+config.addAllowedOrigin("*");`, `+config.setAllowCredentials(true);`}, 1},
		{"spring setAllowedOrigins list star", "A.java", []string{
			`+config.setAllowedOrigins(List.of("*"));`, `+config.addAllowedMethod("POST");`}, 1},
		{"spring explicit allowed origins list", "A.java", []string{
			`+config.setAllowedOrigins(List.of("https://a.example.com"));`, `+config.setAllowCredentials(true);`}, 0},
		{"servlet header star", "A.java", []string{
			`+resp.setHeader("Access-Control-Allow-Origin", "*");`, `+resp.setHeader("Access-Control-Allow-Methods", "POST");`}, 1},
		{"servlet reflect origin", "A.java", []string{
			`+resp.setHeader("Access-Control-Allow-Origin", req.getHeader("Origin"));`}, 1},
		// PHP
		{"php header star", "api/a.php", []string{`+header("Access-Control-Allow-Origin: *");`, `+header("Access-Control-Allow-Methods: POST, DELETE");`}, 1},
		{"php reflect HTTP_ORIGIN", "api/a.php", []string{
			`+header("Access-Control-Allow-Origin: " . $_SERVER['HTTP_ORIGIN']);`,
			`+header("Access-Control-Allow-Credentials: true");`}, 1},
		{"php reflect local", "api/a.php", []string{
			`+$origin = $_SERVER['HTTP_ORIGIN'];`,
			`+header("Access-Control-Allow-Origin: $origin");`,
			`+header("Access-Control-Allow-Credentials: true");`}, 1},
		{"php allowlisted reflect", "api/a.php", []string{
			`+$origin = $_SERVER['HTTP_ORIGIN'];`,
			`+if (in_array($origin, $trusted, true)) {`,
			`+    header("Access-Control-Allow-Origin: $origin");`,
			`+}`}, 0},
		{"php fixed origin", "api/a.php", []string{`+header("Access-Control-Allow-Origin: https://app.example.com");`}, 0},
		{"laravel config star with credentials", "config/cors.php", []string{
			`+    'allowed_origins' => ['*'],`, `+    'supports_credentials' => true,`}, 1},
		{"laravel config explicit", "config/cors.php", []string{
			`+    'allowed_origins' => ['https://a.example.com'],`, `+    'supports_credentials' => true,`}, 0},
		{"laravel header reflect", "app/Http/M.php", []string{
			`+return $next($request)->header('Access-Control-Allow-Origin', $request->header('Origin'));`}, 1},
		{"commented django", "app/settings.py", []string{`+# CORS_ALLOW_ALL_ORIGINS = True`}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(Secrets(fileOf(c.path, 1, c.body...)), finding.CORS)
			if got != c.want {
				t.Fatalf("got %d cors findings, want %d", got, c.want)
			}
		})
	}
}
