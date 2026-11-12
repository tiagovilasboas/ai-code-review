package rules

import "testing"

func TestXSSEscapedBuffersAndOutput(t *testing.T) {
	runXSSCases(t, []xssCase{
		{"go template.HTML of escaped buffer", "web/a.go", []string{
			`+func h(w http.ResponseWriter, r *http.Request) {`,
			`+	var buf bytes.Buffer`,
			`+	buf.WriteString("<p>")`,
			`+	buf.WriteString(html.EscapeString(r.FormValue("q")))`,
			`+	buf.WriteString("</p>")`,
			`+	data := map[string]any{"Body": template.HTML(buf.String())}`,
			`+}`}, nil},
		{"go template.HTML of buffer, Fprintf escaped", "web/a.go", []string{
			`+	var sb strings.Builder`,
			`+	fmt.Fprintf(&sb, "<p>%s</p>", template.HTMLEscapeString(r.FormValue("q")))`,
			`+	return template.HTML(sb.String())`}, nil},
		{"go template.HTML of buffer without a visible write", "web/a.go", []string{
			`+func h() template.HTML {`,
			`+	return template.HTML(buf.String())`,
			`+}`}, []int{2}},
		{"go template.HTML of buffer with a raw write", "web/a.go", []string{
			`+	var buf bytes.Buffer`,
			`+	buf.WriteString(html.EscapeString(a))`,
			`+	buf.WriteString(r.FormValue("q"))`,
			`+	x := template.HTML(buf.String())`}, []int{4}},
		{"go template.HTML of buffer filled by Execute", "web/a.go", []string{
			`+	var buf bytes.Buffer`,
			`+	buf.WriteString("<p>")`,
			`+	tmpl.Execute(&buf, r.FormValue("q"))`,
			`+	x := template.HTML(buf.String())`}, []int{4}},
		{"go template.HTML of request value", "web/a.go", []string{
			`+	x := template.HTML(r.FormValue("q"))`}, []int{1}},
		{"go buffer escaped in another function", "web/a.go", []string{
			`+func a() {`,
			`+	buf.WriteString(html.EscapeString(q))`,
			`+}`,
			`+func b() {`,
			`+	buf.WriteString(r.FormValue("q"))`,
			`+	x := template.HTML(buf.String())`,
			`+}`}, []int{6}},
		{"echo c.HTML concat request", "web/a.go", []string{
			`+	q := c.QueryParam("q")`,
			`+	return c.HTML(http.StatusOK, "<p>"+q+"</p>")`}, []int{2}},
		{"echo c.HTML escaped", "web/a.go", []string{
			`+	q := c.QueryParam("q")`,
			`+	return c.HTML(http.StatusOK, "<p>"+html.EscapeString(q)+"</p>")`}, nil},
		{"gin c.HTML template name", "web/a.go", []string{
			`+	c.HTML(http.StatusOK, "index.tmpl", gin.H{"q": c.Query("q")})`}, nil},
		{"php htmlspecialchars with flags in short echo", "views/a.php", []string{
			`+<h1>Hello <?= htmlspecialchars($_GET['name'], ENT_QUOTES, 'UTF-8') ?></h1>`}, nil},
		{"php htmlspecialchars concatenated with raw", "views/a.php", []string{
			`+<h1>Hello <?= htmlspecialchars($_GET['a'], ENT_QUOTES) . $_GET['b'] ?></h1>`}, []int{1}},
		{"php short echo raw request", "views/a.php", []string{
			`+<h1>Hello <?= $_GET['name'] ?></h1>`}, []int{1}},
		{"blade hyphen text mail view", "resources/views/mail/invoice-text.blade.php", []string{
			`+Hello {!! $name !!}`}, nil},
		{"blade html mail view", "resources/views/mail/invoice.blade.php", []string{
			`+Hello {!! $name !!}`}, []int{1}},
		{"php echo of shell output is not reflection", "public/ping.php", []string{
			`+echo shell_exec("ping -c 1 " . escapeshellarg($_GET['host']));`}, nil},
		{"php echo of shell output inside markup", "public/ping.php", []string{
			`+echo "<pre>" . shell_exec("ping -c 1 " . $_GET['host']) . "</pre>";`}, nil},
		{"php echo of request next to shell output", "public/ping.php", []string{
			`+echo "<pre>" . $_GET['host'] . shell_exec("ls") . "</pre>";`}, []int{1}},
	})
}
