package rules

import "testing"

func TestAuthZCallerInputAndGuards(t *testing.T) {
	cases := []struct {
		name string
		path string
		body []string
		want int
	}{
		{"fastify request.params", "src/r.ts", []string{
			"+app.get('/d/:id', async (request) => {",
			"+  return db.doc.findUnique({ where: { id: request.params.id } });",
			"+});"}, 1},
		{"koa ctx.params", "src/r.js", []string{
			"+router.get('/f/:id', async (ctx) => {",
			"+  ctx.body = await File.findById(ctx.params.id);",
			"+});"}, 1},
		{"next bare params in route file", "app/api/n/[id]/route.ts", []string{
			"+export async function GET(_r: Request, { params }: { params: { id: string } }) {",
			"+  return db.note.findFirst({ where: { id: params.id } });",
			"+}"}, 1},
		{"bare params outside route files is ignored", "src/service/notes.ts", []string{
			"+export async function load(params: { id: string }) {",
			"+  return db.note.findFirst({ where: { id: params.id } });",
			"+}"}, 0},
		{"public flag", "src/r.js", []string{
			"+router.get('/p/:s', async (req, res) => {",
			"+  const p = await Post.findOne({ slug: req.params.s, isPublished: true });",
			"+});"}, 0},
		{"admin role guard", "src/r.js", []string{
			"+router.get('/u/:id', requireRole('admin'), async (req, res) => {",
			"+  res.json(await User.findById(req.params.id));",
			"+});"}, 0},
		{"non-admin role does not exempt", "src/r.js", []string{
			"+router.get('/u/:id', requireRole('member'), async (req, res) => {",
			"+  res.json(await User.findById(req.params.id));",
			"+});"}, 1},
		{"slug mixed with other filters still flagged", "src/r.js", []string{
			"+router.get('/p/:s', async (req, res) => {",
			"+  res.json(await Doc.findOne({ slug: req.params.s, teamId: req.params.t }));",
			"+});"}, 1},
		{"session user in handler", "src/r.js", []string{
			"+router.get('/n/:id', async (req, res) => {",
			"+  const s = await getSession(req); const u = session.user;",
			"+  res.json(await Note.findById(req.params.id));",
			"+});"}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := len(AuthZ(fileOf(c.path, 1, c.body...))); got != c.want {
				t.Fatalf("got %d findings, want %d", got, c.want)
			}
		})
	}
}
