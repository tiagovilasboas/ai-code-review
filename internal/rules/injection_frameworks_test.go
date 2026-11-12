package rules

import "testing"

func TestPathTraversalFrameworks(t *testing.T) {
	const jv = "src/main/java/com/acme/files/DownloadResource.java"
	runFwCases(t, PathTraversal, []fwCase{
		{"jaxrs pathparam", jv, []string{
			` public byte[] download(@PathParam("name") String name) throws IOException {`,
			`+    return Files.readAllBytes(Paths.get(uploadDir, name));`}, []int{2}},
		{"quarkus restpath", jv, []string{
			` public byte[] download(@RestPath String name) throws IOException {`,
			`+    return Files.readAllBytes(Paths.get(uploadDir, name));`}, []int{2}},
		{"jaxrs pathparam with a second annotation", jv, []string{
			` public byte[] download(@PathParam("name") @Parameter(description = "file (relative)") String name) {`,
			`+    return Files.readAllBytes(Paths.get(uploadDir, name));`}, []int{2}},
		{"quarkus restquery", jv, []string{
			` public byte[] download(@RestQuery("f") String f) throws IOException {`,
			`+    return Files.readAllBytes(Paths.get(uploadDir, f));`}, []int{2}},
		{"jaxrs normalize and prefix check", jv, []string{
			` public byte[] download(@RestPath String name) throws IOException {`,
			`+    java.nio.file.Path p = Paths.get(uploadDir).resolve(name).normalize();`,
			`+    if (!p.startsWith(Paths.get(uploadDir))) { throw new ForbiddenException(); }`,
			`+    return Files.readAllBytes(p);`}, nil},
		{"jaxrs filename utils", jv, []string{
			` public byte[] download(@RestPath String name) throws IOException {`,
			`+    return Files.readAllBytes(Paths.get(uploadDir, FilenameUtils.getName(name)));`}, nil},
		{"ci4 readfile writepath getGet", "app/Controllers/Download.php", []string{
			`+    readfile(WRITEPATH . 'uploads/' . $this->request->getGet('file'));`}, []int{1}},
		{"ci4 readfile via getVar", "app/Controllers/Download.php", []string{
			`+    $f = $this->request->getVar('file');`,
			`+    return file_get_contents(WRITEPATH . 'uploads/' . $f);`}, []int{2}},
		{"ci4 basename", "app/Controllers/Download.php", []string{
			`+    $name = basename($this->request->getGet('file'));`,
			`+    readfile(WRITEPATH . 'uploads/' . $name);`}, nil},
		{"aiohttp match_info open", "app/handlers.py", []string{
			`+    name = request.match_info["name"]`,
			`+    with open(os.path.join(UPLOAD_DIR, name), "rb") as fh:`}, []int{2}},
		{"aiohttp basename", "app/handlers.py", []string{
			`+    name = os.path.basename(request.match_info["name"])`,
			`+    with open(os.path.join(UPLOAD_DIR, name), "rb") as fh:`}, nil},
		{"aiohttp realpath prefix", "app/handlers.py", []string{
			`+    name = request.match_info["name"]`,
			`+    p = os.path.realpath(os.path.join(UPLOAD_DIR, name))`,
			`+    if not p.startswith(UPLOAD_DIR): raise web.HTTPForbidden()`,
			`+    with open(p, "rb") as fh:`}, nil},
		{"go 1.22 PathValue", "internal/files/handler.go", []string{
			`+	data, err := os.ReadFile(filepath.Join(h.root, r.PathValue("name")))`}, []int{1}},
		{"go 1.22 PathValue via variable", "internal/files/handler.go", []string{
			`+	name := r.PathValue("name")`,
			`+	f, err := os.Open(filepath.Join(h.root, name))`}, []int{2}},
		{"go PathValue Base", "internal/files/handler.go", []string{
			`+	name := filepath.Base(r.PathValue("name"))`,
			`+	data, err := os.ReadFile(filepath.Join(h.root, name))`}, nil},
	})
}

func TestSQLInjectionJdbcTemplate(t *testing.T) {
	const jv = "src/main/java/com/acme/customers/CustomerDao.java"
	runFwCases(t, SQLInjection, []fwCase{
		{"jdbcTemplate query concat", jv, []string{
			` public List<Customer> find(String lastName) {`,
			`+    return jdbcTemplate.query("SELECT * FROM customers WHERE last_name = '" + lastName + "'", rowMapper);`}, []int{2}},
		{"jdbcTemplate update concat", jv, []string{
			`+    return jdbcTemplate.update("UPDATE customers SET last_name = '" + lastName + "' WHERE id = ?", id);`}, []int{1}},
		{"jdbc batchUpdate concat", jv, []string{
			`+    namedJdbc.batchUpdate("DELETE FROM customers WHERE last_name = '" + lastName + "'");`}, []int{1}},
		{"jdbcTemplate queryForObject concat", jv, []string{
			`+    return jdbcTemplate.queryForObject("SELECT count(*) FROM customers WHERE last_name = '" + lastName + "'", Integer.class);`}, []int{1}},
		{"jdbcTemplate queryForRowSet concat", jv, []string{
			`+    return jdbcTemplate.queryForRowSet("SELECT * FROM customers WHERE last_name = '" + lastName + "'");`}, []int{1}},
		{"jdbcTemplate placeholders", jv, []string{
			`+    jdbcTemplate.query("SELECT * FROM customers WHERE last_name = ?", rowMapper, lastName);`,
			`+    jdbcTemplate.update("UPDATE customers SET last_name = ? WHERE id = ?", lastName, id);`}, nil},
		{"update on a non jdbc object", jv, []string{
			`+    service.update("name = " + lastName);`}, nil},
	})
}

func TestCmdInjectionJavaShell(t *testing.T) {
	const jv = "src/main/java/com/acme/net/PingService.java"
	runFwCases(t, CmdInjection, []fwCase{
		{"processbuilder sh -c concat", jv, []string{
			`+    new ProcessBuilder("sh", "-c", "ping -c 1 " + host).start();`}, []int{1}},
		{"runtime exec array", jv, []string{
			`+    Runtime.getRuntime().exec(new String[]{"sh", "-c", "ping -c 1 " + host});`}, []int{1}},
		{"runtime variable exec array", jv, []string{
			`+    runtime.exec(new String[] {"/bin/sh", "-c", "ping -c 1 " + host});`}, []int{1}},
		{"processbuilder Arrays.asList", jv, []string{
			`+    new ProcessBuilder(Arrays.asList("bash", "-c", "ping -c 1 " + host)).start();`}, []int{1}},
		{"processbuilder List.of", jv, []string{
			`+    new ProcessBuilder(List.of("sh", "-c", "ping -c 1 " + host)).start();`}, []int{1}},
		{"processbuilder command()", jv, []string{
			`+    pb.command("sh", "-c", "ping -c 1 " + host);`}, []int{1}},
		{"processbuilder argv", jv, []string{
			`+    new ProcessBuilder("ping", "-c", "1", host).start();`}, nil},
		{"processbuilder list argv", jv, []string{
			`+    new ProcessBuilder(List.of("ping", "-c", "1", host)).start();`}, nil},
		{"command() argv", jv, []string{
			`+    pb.command("ping", "-c", "1", host);`}, nil},
		{"runtime exec array argv", jv, []string{
			`+    Runtime.getRuntime().exec(new String[]{"ping", "-c", "1", host});`}, nil},
		{"shell with constant script", jv, []string{
			`+    new ProcessBuilder("sh", "-c", "ping -c 1 localhost").start();`}, nil},
	})
}
