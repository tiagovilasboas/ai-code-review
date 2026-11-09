package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestAuthZFrameworkHandlers(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		{"gin repo lookup", "h.go", []string{
			"+func Get(c *gin.Context) {",
			`+	id := c.Param("id")`,
			"+	a, _ := repo.GetAccount(id)",
			"+	c.JSON(200, a)",
			"+}"}, 1},
		{"gin inline param", "h.go", []string{
			"+func Get(c *gin.Context) {",
			`+	a, _ := repo.GetAccount(c.Param("id"))`,
			"+	c.JSON(200, a)",
			"+}"}, 1},
		{"gin closure handler", "h.go", []string{
			`+r.DELETE("/a/:id", func(c *gin.Context) {`,
			`+	id := c.Param("id")`,
			"+	db.Delete(&Account{}, id)",
			"+})"}, 1},
		{"gin user from context", "h.go", []string{
			"+func Get(c *gin.Context) {",
			`+	id := c.Param("id")`,
			`+	uid := c.GetString("userID")`,
			"+	a, _ := repo.GetForUser(id, uid)",
			"+}"}, 0},
		{"gin slug param", "h.go", []string{
			"+func Page(c *gin.Context) {",
			`+	slug := c.Param("slug")`,
			"+	p, _ := repo.GetPage(slug)",
			"+}"}, 0},
		{"gin param never reaches a lookup", "h.go", []string{
			"+func Echo(c *gin.Context) {",
			`+	id := c.Param("id")`,
			`+	c.JSON(200, gin.H{"id": id})`,
			"+}"}, 0},
		{"gin go test file", "h_test.go", []string{
			"+func Get(c *gin.Context) {",
			`+	id := c.Param("id")`,
			"+	a, _ := repo.GetAccount(id)",
			"+}"}, 0},
		{"gin repo call is not a principal", "h.go", []string{
			"+func Get(c *gin.Context) {",
			`+	id := c.Param("id")`,
			"+	a, _ := h.store.Get(id)",
			"+}"}, 1},
		{"nest body dto id", "c.ts", []string{
			"+@Put()",
			"+replace(@Body() dto: ReplaceDto) {",
			"+  return this.svc.update(dto.id, dto);",
			"+}"}, 1},
		{"nest create without id", "c.ts", []string{
			"+@Post()",
			"+create(@Body() dto: CreateDto) {",
			"+  return this.svc.create(dto);",
			"+}"}, 0},
		{"nest req.user", "c.ts", []string{
			"+@Get(':id')",
			"+one(@Param('id') id: string, @Req() req) {",
			"+  return this.svc.findOne(id, req.user.id);",
			"+}"}, 0},
		{"nest class-level policy guard", "c.ts", []string{
			"+@UseGuards(JwtGuard, PolicyGuard)",
			"+@Controller('x')",
			"+export class XController {",
			"+  @Get(':id')",
			"+  one(@Param('id') id: string) {",
			"+    return this.svc.findOne(id);",
			"+  }",
			"+}"}, 0},
		{"nest auth guard only is not ownership", "c.ts", []string{
			"+@Get(':id')",
			"+@UseGuards(AuthGuard('jwt'))",
			"+one(@Param('id') id: string) {",
			"+  return this.svc.findOne(id);",
			"+}"}, 1},
		{"nest public route", "c.ts", []string{
			"+@Public()",
			"+@Get(':id')",
			"+one(@Param('id') id: string) {",
			"+  return this.svc.findOne(id);",
			"+}"}, 0},
		{"nest slug param", "c.ts", []string{
			"+@Get(':slug')",
			"+one(@Param('slug') slug: string) {",
			"+  return this.svc.findBySlug(slug);",
			"+}"}, 0},
		{"flask login_required alone", "v.py", []string{
			`+@bp.route("/d/<int:doc_id>")`,
			"+@login_required",
			"+def show(doc_id):",
			"+    d = Doc.query.get_or_404(doc_id)",
			"+    return jsonify(d.to_dict())"}, 1},
		{"flask user scoped filter", "v.py", []string{
			`+@bp.route("/d/<int:doc_id>")`,
			"+def show(doc_id):",
			"+    d = Doc.query.filter_by(id=doc_id, user_id=current_user.id).first_or_404()",
			"+    return jsonify(d.to_dict())"}, 0},
		{"flask g.user", "v.py", []string{
			`+@bp.get("/d/<doc_id>")`,
			"+def show(doc_id):",
			"+    d = Doc.query.get(doc_id)",
			"+    if d.owner != g.user:",
			"+        abort(404)"}, 0},
		{"flask second view is not shadowed by first guard", "v.py", []string{
			`+@bp.get("/a/<a_id>")`,
			"+def a(a_id):",
			"+    x = A.query.get(a_id)",
			"+    if x.owner_id != current_user.id:",
			"+        abort(403)",
			`+@bp.get("/b/<b_id>")`,
			"+def b(b_id):",
			"+    return B.query.get(b_id).to_dict()"}, 1},
		{"flask no url id", "v.py", []string{
			`+@bp.get("/mine")`,
			"+def mine():",
			"+    return Doc.query.get(request.args['id']).to_dict()"}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(AuthZ(fileOf("src/"+c.path, 1, c.body...)), finding.AuthZ)
			if got != c.want {
				t.Fatalf("got %d authz findings, want %d", got, c.want)
			}
		})
	}
}
