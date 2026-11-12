package rules

import "testing"

func TestXSSFastHTTPAndVertx(t *testing.T) {
	const gof = "internal/web/search.go"
	const jv = "src/main/java/com/acme/web/HelloHandler.java"
	runFwCases(t, XSS, []fwCase{
		{"fasthttp fprintf peek", gof, []string{
			` func h(ctx *fasthttp.RequestCtx) {`,
			`	q := ctx.QueryArgs().Peek("q")`,
			`+	fmt.Fprintf(ctx, "<h2>Results for %s</h2>", q)`}, []int{3}},
		{"fasthttp post args write string", gof, []string{
			`	n := ctx.PostArgs().Peek("name")`,
			`+	ctx.WriteString("<b>" + string(n) + "</b>")`}, []int{2}},
		{"fasthttp set body string sprintf", gof, []string{
			`	n := ctx.FormValue("name")`,
			`+	ctx.SetBodyString(fmt.Sprintf("<p>%s</p>", n))`}, []int{2}},
		{"fasthttp user value", gof, []string{
			`	id := ctx.UserValue("id")`,
			`+	fmt.Fprintf(ctx, "<i>%v</i>", id)`}, []int{2}},
		{"fasthttp escaped", gof, []string{
			`	q := ctx.QueryArgs().Peek("q")`,
			`+	fmt.Fprintf(ctx, "<h2>%s</h2>", html.EscapeString(string(q)))`}, nil},
		{"fasthttp plain text content type", gof, []string{
			`	q := ctx.QueryArgs().Peek("q")`,
			`	ctx.SetContentType("text/plain")`,
			`+	fmt.Fprintf(ctx, "<h2>%s</h2>", q)`}, nil},
		{"fasthttp json content type", gof, []string{
			`	q := ctx.QueryArgs().Peek("q")`,
			`	ctx.SetContentType("application/json")`,
			`+	ctx.SetBodyString("<" + string(q) + ">")`}, nil},
		{"fasthttp constant html", gof, []string{
			`+	fmt.Fprintf(ctx, "<h2>Results</h2>")`}, nil},
		{"vertx end concat", jv, []string{
			`+    String name = ctx.request().getParam("name");`,
			`+    ctx.response().putHeader("Content-Type", "text/html").end("<h1>Hello " + name + "</h1>");`}, []int{2}},
		{"vertx path param", jv, []string{
			`+    String id = ctx.pathParam("id");`,
			`+    ctx.response().end("<p>" + id + "</p>");`}, []int{2}},
		{"vertx query param status chain", jv, []string{
			`+    String q = rc.queryParam("q").get(0);`,
			`+    rc.response().setStatusCode(200).end("<p>" + q + "</p>");`}, []int{2}},
		{"vertx guava escaper", jv, []string{
			`+    String name = ctx.request().getParam("name");`,
			`+    ctx.response().end("<h1>" + HtmlEscapers.htmlEscaper().escape(name) + "</h1>");`}, nil},
		{"vertx commons escape", jv, []string{
			`+    String name = ctx.request().getParam("name");`,
			`+    ctx.response().end("<h1>" + StringEscapeUtils.escapeHtml4(name) + "</h1>");`}, nil},
		{"vertx json body", jv, []string{
			`+    String name = ctx.request().getParam("name");`,
			`+    ctx.response().putHeader("content-type", "application/json").end("{\"n\": \"" + name + "\"}");`}, nil},
	})
}
