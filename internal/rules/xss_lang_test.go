package rules

import "testing"

type xssCase struct {
	name, path string
	body       []string
	want       []int
}

func runXSSCases(t *testing.T, cases []xssCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantLineNums(t, xssLines(t, c.path, c.body...), c.want...)
		})
	}
}

func TestXSSPythonResponses(t *testing.T) {
	runXSSCases(t, []xssCase{
		{"flask f-string return", "app/v.py", []string{`+q = request.args.get("q")`, `+return f"<h1>Results for {q}</h1>"`}, []int{2}},
		{"django HttpResponse concat", "app/v.py", []string{`+n = request.GET["name"]`, `+return HttpResponse("<b>" + n + "</b>")`}, []int{2}},
		{"fastapi HTMLResponse inline", "app/v.py", []string{`+return HTMLResponse(content=f"<p>{request.query_params['q']}</p>")`}, []int{1}},
		{"flask Markup of dynamic", "app/v.py", []string{`+bio = Markup(user_bio)`}, []int{1}},
		{"escaped value", "app/v.py", []string{`+q = request.args.get("q")`, `+return f"<h1>{escape(q)}</h1>"`}, nil},
		{"escaped through local", "app/v.py", []string{`+q = html.escape(request.args.get("q"))`, `+return f"<h1>{q}</h1>"`}, nil},
		{"no markup", "app/v.py", []string{`+q = request.args.get("q")`, `+return f"results: {q}"`}, nil},
		{"jsonify", "app/v.py", []string{`+q = request.args.get("q")`, `+return jsonify(html=f"<b>{q}</b>")`}, nil},
		{"plain mimetype", "app/v.py", []string{`+q = request.args.get("q")`, `+return Response(f"<b>{q}</b>", mimetype="text/plain")`}, nil},
		{"constant markup", "app/v.py", []string{`+return "<h1>Hello</h1>"`}, nil},
		{"Markup literal", "app/v.py", []string{`+sep = Markup("<br>")`}, nil},
		{"jinja autoescape false", "t/a.html", []string{`+{% autoescape false %}`}, []int{1}},
		{"jinja safe filter", "t/a.html", []string{`+<div>{{ post.body|safe }}</div>`}, []int{1}},
	})
}

func TestXSSGoResponses(t *testing.T) {
	runXSSCases(t, []xssCase{
		{"Fprintf format arg", "h/a.go", []string{`+name := r.URL.Query().Get("name")`, `+fmt.Fprintf(w, "<h1>Hello %s</h1>", name)`}, []int{2}},
		{"Fprintf inline input", "h/a.go", []string{`+fmt.Fprintf(w, "<p>%s</p>", r.FormValue("q"))`}, []int{1}},
		{"w.Write concat", "h/a.go", []string{`+n := r.FormValue("n")`, `+w.Write([]byte("<b>" + n + "</b>"))`}, []int{2}},
		{"Sprintf then write", "h/a.go", []string{`+page := fmt.Sprintf("<h1>%s</h1>", r.FormValue("q"))`, `+io.WriteString(w, page)`}, []int{2}},
		{"gin html Data", "h/a.go", []string{`+q := c.Query("q")`, `+c.Data(200, "text/html; charset=utf-8", []byte("<p>"+q+"</p>"))`}, []int{2}},
		{"template.HTML dynamic", "h/a.go", []string{`+data := map[string]any{"bio": template.HTML(user.Bio)}`}, []int{1}},
		{"escaped format arg", "h/a.go", []string{`+n := r.FormValue("n")`, `+fmt.Fprintf(w, "<h1>%s</h1>", html.EscapeString(n))`}, nil},
		{"escaped through local", "h/a.go", []string{`+n := html.EscapeString(r.FormValue("n"))`, `+fmt.Fprintf(w, "<h1>%s</h1>", n)`}, nil},
		{"plain text no markup", "h/a.go", []string{`+fmt.Fprintf(w, "hello %s", r.FormValue("n"))`}, nil},
		{"constant markup", "h/a.go", []string{`+fmt.Fprintf(w, "<h1>%s</h1>", title)`}, nil},
		{"json content type", "h/a.go", []string{`+w.Header().Set("Content-Type", "application/json")`, `+fmt.Fprintf(w, "{\"h\":\"<b>%s</b>\"}", r.FormValue("n"))`}, nil},
		{"template.HTML literal", "h/a.go", []string{`+x := template.HTML("<br>")`}, nil},
		{"template.HTML sanitized", "h/a.go", []string{`+x := template.HTML(policy.Sanitize(user.Bio))`}, nil},
		{"fprintf to buffer", "h/a.go", []string{`+fmt.Fprintf(&buf, "<p>%s</p>", r.FormValue("q"))`}, nil},
	})
}

func TestXSSJavaResponses(t *testing.T) {
	runXSSCases(t, []xssCase{
		{"servlet println concat", "web/S.java", []string{`+String n = request.getParameter("name");`, `+out.println("<h1>Hello " + n + "</h1>");`}, []int{2}},
		{"servlet req parameter inline", "web/S.java", []string{`+resp.getWriter().println("<h1>" + req.getParameter("q") + "</h1>");`}, []int{1}},
		{"servlet getWriter inline", "web/S.java", []string{`+response.getWriter().write("<p>" + request.getParameter("q") + "</p>");`}, []int{1}},
		{"servlet printf", "web/S.java", []string{`+String n = request.getParameter("name");`, `+out.printf("<b>%s</b>", n);`}, []int{2}},
		{"spring return html", "web/C.java", []string{`+public String hi(@RequestParam String name) {`, `+    return "<h1>" + name + "</h1>";`}, []int{2}},
		{"escaped", "web/S.java", []string{`+String n = request.getParameter("name");`, `+out.println("<h1>" + StringEscapeUtils.escapeHtml4(n) + "</h1>");`}, nil},
		{"escaped through local", "web/S.java", []string{`+String n = HtmlUtils.htmlEscape(request.getParameter("name"));`, `+out.println("<h1>" + n + "</h1>");`}, nil},
		{"json content type", "web/S.java", []string{`+response.setContentType("application/json");`, `+out.println("{\"h\":\"<b>" + request.getParameter("q") + "</b>\"}");`}, nil},
		{"constant", "web/S.java", []string{`+out.println("<h1>Hello</h1>");`}, nil},
		{"jsp scriptlet output", "web/a.jsp", []string{`+<p><%= request.getParameter("q") %></p>`}, []int{1}},
		{"jsp param EL", "web/a.jsp", []string{`+<input value="${param.q}">`}, []int{1}},
		{"jsp c:out escapeXml false", "web/a.jsp", []string{`+<c:out value="${comment.body}" escapeXml="false"/>`}, []int{1}},
		{"jsp c:out default", "web/a.jsp", []string{`+<c:out value="${param.q}"/>`}, nil},
		{"jsp escapeXml fn", "web/a.jsp", []string{`+<p>${fn:escapeXml(param.q)}</p>`}, nil},
		{"jsp c:if on param", "web/a.jsp", []string{`+<c:if test="${not empty param.q}">`}, nil},
		{"jsp escaped scriptlet", "web/a.jsp", []string{`+<%= Encode.forHtml(request.getParameter("q")) %>`}, nil},
		{"thymeleaf utext", "t/a.html", []string{`+<div th:utext="${comment.body}"></div>`}, []int{1}},
		{"thymeleaf inline unescaped", "t/a.html", []string{`+<p>[(${comment.body})]</p>`}, []int{1}},
		{"thymeleaf text", "t/a.html", []string{`+<div th:text="${comment.body}"></div>`}, nil},
		{"thymeleaf utext message key", "t/a.html", []string{`+<div th:utext="#{home.welcome}"></div>`}, nil},
	})
}

func TestXSSPHPResponses(t *testing.T) {
	runXSSCases(t, []xssCase{
		{"echo superglobal", "p/a.php", []string{`+echo $_GET['q'];`}, []int{1}},
		{"int cast assignment then echo", "p/a.php", []string{`+$id = (int) $_GET['id'];`, `+echo "<p>Item #" . $id . "</p>";`}, nil},
		{"echo interpolation", "p/a.php", []string{`+$n = $_GET['name'];`, `+echo "<h1>Hello $n</h1>";`}, []int{2}},
		{"echo concat", "p/a.php", []string{`+echo "<b>" . $_POST['n'] . "</b>";`}, []int{1}},
		{"short echo tag", "p/a.php", []string{`+<p><?= $_GET['q'] ?></p>`}, []int{1}},
		{"laravel input echo", "p/a.php", []string{`+$q = $request->input('q');`, `+echo "<p>{$q}</p>";`}, []int{2}},
		{"printf", "p/a.php", []string{`+printf("<b>%s</b>", $_GET['n']);`}, []int{1}},
		{"htmlspecialchars", "p/a.php", []string{`+echo "<b>" . htmlspecialchars($_GET['n']) . "</b>";`}, nil},
		{"escaped through local", "p/a.php", []string{`+$n = htmlspecialchars($_GET['n'], ENT_QUOTES);`, `+echo "<b>$n</b>";`}, nil},
		{"int cast", "p/a.php", []string{`+echo "<b>" . (int)$_GET['id'] . "</b>";`}, nil},
		{"intval", "p/a.php", []string{`+echo intval($_GET['id']);`}, nil},
		{"echo of command output", "p/a.php", []string{`+$f = $_GET['f'];`, `+$out = shell_exec('convert ' . escapeshellcmd($f));`, `+echo $out;`}, nil},
		{"constant echo", "p/a.php", []string{`+echo "<h1>Hello</h1>";`}, nil},
		{"unrelated variable", "p/a.php", []string{`+echo "<h1>$title</h1>";`}, nil},
		{"json header", "p/a.php", []string{`+header('Content-Type: application/json');`, `+echo $_GET['q'];`}, nil},
		{"blade raw variable", "r/a.blade.php", []string{`+<div>{!! $post->body !!}</div>`}, []int{1}},
		{"blade escaped", "r/a.blade.php", []string{`+<div>{!! nl2br(e($post->body)) !!}</div>`}, nil},
		{"blade default escaping", "r/a.blade.php", []string{`+<div>{{ $post->body }}</div>`}, nil},
		{"blade helper", "r/a.blade.php", []string{`+{!! csrf_field() !!}`, `+{!! __('messages.welcome') !!}`, `+{!! $slot !!}`}, nil},
		{"twig raw", "t/a.html.twig", []string{`+<div>{{ comment.body|raw }}</div>`}, []int{1}},
		{"twig escape then raw", "t/a.html.twig", []string{`+<div>{{ comment.body|escape|nl2br|raw }}</div>`}, nil},
		{"twig default", "t/a.html.twig", []string{`+<div>{{ comment.body }}</div>`}, nil},
		{"twig autoescape false", "t/a.html.twig", []string{`+{% autoescape false %}`}, []int{1}},
		{"twig literal raw", "t/a.html.twig", []string{`+{{ '<br>'|raw }}`}, nil},
	})
}
