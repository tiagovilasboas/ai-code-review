package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestPathTraversalFires(t *testing.T) {
	runRuleCases(t, PathTraversal, finding.PathTraversal, "CWE-22", []ruleCase{
		{"js sendFile from req.params", mkFile("src/api/files.ts",
			`+  res.sendFile(path.join(__dirname, "uploads", req.params.name));`,
		), []int{1}, finding.High},
		{"js readFile via joined var", mkFile("src/api/files.js",
			`+  const file = path.join(UPLOAD_DIR, req.query.file);`,
			`+  const data = await fs.promises.readFile(file, "utf8");`,
		), []int{2}, finding.High},
		{"js writeFile destructured body", mkFile("src/api/files.js",
			`+  const { filename, content } = req.body;`,
			`+  fs.writeFileSync("/srv/data/" + filename, content);`,
		), []int{2}, finding.High},
		{"python open from request.args", mkFile("app/files.py",
			`+    name = request.args.get("name")`,
			`+    with open(os.path.join(BASE_DIR, name)) as fh:`,
		), []int{2}, finding.High},
		{"python send_file", mkFile("app/files.py",
			`+    return send_file("/srv/reports/" + request.args["report"])`,
		), []int{1}, finding.High},
		{"go os.Open from query", mkFile("internal/files/handler.go",
			`+	name := r.URL.Query().Get("name")`,
			`+	f, err := os.Open(filepath.Join(baseDir, name))`,
		), []int{2}, finding.High},
		{"go ServeFile third arg", mkFile("internal/files/handler.go",
			`+	http.ServeFile(w, r, "/srv/files/"+r.FormValue("f"))`,
		), []int{1}, finding.High},
		{"java new File with RequestParam", mkFile("src/main/java/FileController.java",
			`+    public byte[] get(@RequestParam("name") String name) throws IOException {`,
			`+        return Files.readAllBytes(new File(baseDir, name).toPath());`,
		), []int{2}, finding.High},
		{"php file_get_contents $_GET", mkFile("public/view.php",
			`+echo file_get_contents("/var/data/" . $_GET['file']);`,
		), []int{1}, finding.High},
		{"php include bare", mkFile("public/index.php",
			`+include $_GET['page'] . '.php';`,
		), []int{1}, finding.High},
		{"guard in a sibling handler does not count", mkFile("src/api/files.ts",
			`+export function preview(req: Request, res: Response) {`,
			`+  const target = path.resolve(DIR, req.params.name);`,
			`+  if (!target.startsWith(DIR + path.sep)) return res.sendStatus(400);`,
			`+  return res.sendFile(target);`,
			`+}`,
			`+export function download(req: Request, res: Response) {`,
			`+  return res.sendFile(path.join(DIR, req.params.name));`,
			`+}`,
		), []int{7}, finding.High},
	})
}

func TestPathTraversalTrueNegatives(t *testing.T) {
	runRuleCases(t, PathTraversal, finding.PathTraversal, "CWE-22", []ruleCase{
		{"js fixed path", mkFile("src/api/files.ts",
			`+  const data = fs.readFileSync(path.join(__dirname, "config.json"));`,
		), nil, ""},
		{"js path not from request", mkFile("src/jobs/files.ts",
			`+  const data = fs.readFileSync(cfg.reportPath);`,
		), nil, ""},
		{"js resolve plus prefix check", mkFile("src/api/files.ts",
			`+  const target = path.resolve(ROOT, req.params.name);`,
			`+  if (!target.startsWith(ROOT + path.sep)) return res.sendStatus(400);`,
			`+  res.sendFile(target);`,
		), nil, ""},
		{"js sendFile with root option", mkFile("src/api/files.ts",
			`+  res.sendFile(req.params.name, { root: PUBLIC_DIR });`,
		), nil, ""},
		{"js basename", mkFile("src/api/files.ts",
			`+  const name = path.basename(req.query.file);`,
			`+  fs.createReadStream(path.join(DIR, name)).pipe(res);`,
		), nil, ""},
		{"js comment", mkFile("src/api/files.ts",
			`+  // fs.readFileSync(req.query.file)`,
		), nil, ""},
		{"python secure_filename", mkFile("app/files.py",
			`+    name = secure_filename(request.args["name"])`,
			`+    return send_file(os.path.join(UPLOADS, name))`,
		), nil, ""},
		{"python realpath plus commonpath", mkFile("app/files.py",
			`+    target = os.path.realpath(os.path.join(BASE, request.args["name"]))`,
			`+    if os.path.commonpath([BASE, target]) != BASE:`,
			`+        abort(400)`,
			`+    return open(target).read()`,
		), nil, ""},
		{"go clean plus HasPrefix", mkFile("internal/files/handler.go",
			`+	p := filepath.Clean(filepath.Join(base, r.URL.Query().Get("f")))`,
			`+	if !strings.HasPrefix(p, base+string(os.PathSeparator)) {`,
			`+		return`,
			`+	}`,
			`+	data, err := os.ReadFile(p)`,
		), nil, ""},
		{"java canonical path check", mkFile("src/main/java/FileController.java",
			`+        File f = new File(baseDir, request.getParameter("name")).getCanonicalFile();`,
			`+        String canonical = f.getCanonicalPath();`,
			`+        if (!canonical.startsWith(baseDir.getCanonicalPath())) throw new SecurityException();`,
		), nil, ""},
		{"php basename", mkFile("public/view.php",
			`+$name = basename($_GET['file']);`,
			`+echo file_get_contents("/var/data/" . $name);`,
		), nil, ""},
		{"php literal include", mkFile("public/index.php",
			`+require_once 'config.php';`,
		), nil, ""},
	})
}

func TestPathFromStoredRecordIsSilent(t *testing.T) {
	f := fileOf("src/routes/files.js", 1,
		"+router.get('/files/:id', async (req, res) => {",
		"+  const file = await File.findById(req.params.id);",
		"+  res.sendFile(file.path);",
		"+});",
	)
	if got := PathTraversal(f); len(got) != 0 {
		t.Fatalf("a path read from a stored record is not request input, got %+v", got)
	}
}

func TestPathTraversalSpringResources(t *testing.T) {
	runRuleCases(t, PathTraversal, finding.PathTraversal, "CWE-22", []ruleCase{
		{"urlresource from resolved request name", mkFile("src/main/java/FileController.java",
			`+    public Resource get(@RequestParam("filename") String filename) throws Exception {`,
			`+        Path file = Paths.get(base).resolve(filename);`,
			`+        return new UrlResource(file.toUri());`,
		), []int{3}, finding.High},
		{"file system resource from path variable", mkFile("src/main/java/FileController.java",
			`+    public Resource get(@PathVariable String name) throws Exception {`,
			`+        Path file = baseDir.resolve(name);`,
			`+        return new FileSystemResource(file);`,
		), []int{3}, finding.High},
	})
	runRuleCases(t, PathTraversal, finding.PathTraversal, "CWE-22", []ruleCase{
		{"urlresource after normalize and startsWith", mkFile("src/main/java/FileController.java",
			`+    public Resource get(@RequestParam("filename") String filename) throws Exception {`,
			`+        Path file = baseDir.resolve(filename).normalize();`,
			`+        if (!file.startsWith(baseDir)) { throw new IllegalArgumentException(); }`,
			`+        return new UrlResource(file.toUri());`,
		), nil, finding.High},
	})
}
