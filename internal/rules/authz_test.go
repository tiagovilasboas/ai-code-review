package rules

import (
	"reflect"
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestAuthZSamplePR(t *testing.T) {
	f := fileOf("src/api/orders.ts", 30,
		"   return res.json(orders);",
		" }",
		" ",
		"+export async function getOrder(req: Request, res: Response) {",
		"+  const orderId = req.params.id;",
		"+  // IDOR smoke target: identifier from path used without ownership/tenant check in this hunk",
		"+  const order = await db.orders.findById(orderId);",
		"+  if (!order) {",
		`+    return res.status(404).json({ error: "not_found" });`,
		"+  }",
		"+  return res.json(order);",
		"+}",
		"+",
		" export async function createOrder(req: Request, res: Response) {",
	)
	got := AuthZ(f)
	wantLines(t, got, 36)
	want := finding.Finding{
		Title: "Missing object-level authorization", Class: finding.AuthZ,
		Path: "src/api/orders.ts", Line: 36, CWE: "CWE-639",
		OWASP:    []string{"A01:2021", "API1:2023", "ASVS-5.0-8.2.2"},
		Severity: finding.High, Action: finding.Block,
	}
	if !reflect.DeepEqual(got[0], want) {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
}

func TestAuthZ(t *testing.T) {
	cases := []struct {
		name string
		body []string
		want []int
	}{
		{"unrelated userId in handler does not silence", []string{
			"export async function getOrder(req: Request, res: Response) {",
			"  const userId = req.params.userId; // analytics only",
			"  const order = await db.orders.findById(req.params.id);",
			"  return res.json(order);",
			"}",
		}, []int{4}},
		{"userId on sink line outside the query does not silence", []string{
			"export async function getOrder(req: Request, res: Response) {",
			"  const order = await db.orders.findUnique({ where: { id: req.params.id } }); const userId = 1;",
			"  return res.json(order);",
			"}",
		}, []int{3}},
		{"check in another handler does not cover this one", []string{
			"export async function updateOrder(req: Request, res: Response) {",
			"  await assertOwner(req.user, req.params.id);",
			"}",
			"export async function getOrder(req: Request, res: Response) {",
			"  const order = await db.orders.findById(req.params.id);",
			"  return res.json(order);",
			"}",
		}, []int{6}},
		{"ownership filter in the query is silence", []string{
			"export async function getOrder(req: Request, res: Response) {",
			"  const order = await db.orders.findFirst({",
			"    where: { id: req.params.id, ownerId: req.user.id },",
			"  });",
			"  return res.json(order);",
			"}",
		}, nil},
		{"principal check in same handler is silence", []string{
			"export async function getOrder(req: Request, res: Response) {",
			"  const order = await db.orders.findById(req.params.id);",
			"  if (order.ownerId !== req.user.id) return res.sendStatus(403);",
			"  return res.json(order);",
			"}",
		}, nil},
		{"principal check outside the try block is silence", []string{
			"export async function getOrder(",
			"  req: Request,",
			"  res: Response,",
			") {",
			"  const viewer = req.user;",
			"  try {",
			"    const order = await db.orders.findById(req.params.id);",
			"    if (order.ownerId !== viewer.id) return res.sendStatus(403);",
			"    return res.json(order);",
			"  } catch (err) {",
			"    return res.sendStatus(500);",
			"  }",
			"}",
		}, nil},
		{"check in next handler does not leak back", []string{
			`router.get("/orders/:id", async (req, res) => {`,
			"  const order = await db.orders.findById(req.params.id);",
			"  res.json(order);",
			"});",
			`router.put("/orders/:id", async (req, res) => {`,
			"  if (!canAccess(req.user, req.params.id)) return res.sendStatus(403);",
			"});",
		}, []int{3}},
		{"no caller id in the diff is silence", []string{
			"export async function getOrder(id: string) {",
			"  return db.orders.findById(id);",
			"}",
		}, nil},
		{"comment-only caller id is silence", []string{
			"// uses req.params.id somewhere",
			"export async function getOrder(id: string) {",
			"  return db.orders.findById(id);",
			"}",
		}, nil},
		{"string with braces does not skew handler scope", []string{
			"export async function getOrder(req: Request, res: Response) {",
			`  const msg = "}}}";`,
			"  const order = await db.orders.findById(req.params.id);",
			"  return res.json(order);",
			"}",
		}, []int{4}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantLines(t, AuthZ(handlerFile(c.body...)), c.want...)
		})
	}
}

func TestAuthZIgnoresContextSink(t *testing.T) {
	f := fileOf("a.ts", 1,
		" const o = await db.findById(req.params.id);",
		"+const x = req.params.id;",
	)
	wantLines(t, AuthZ(f))
}

func TestStripStrings(t *testing.T) {
	cases := map[string]string{
		`a "b{" c`:            `a "" c`,
		`'it\'s' + {`:         `"" + {`,
		`x = "unterminated {`: `x = "unterminated {`,
		"`t{}` + '}'":         `"" + ""`,
		`'abc\'`:              `""`,
	}
	for in, want := range cases {
		if got := stripStrings(in); got != want {
			t.Errorf("stripStrings(%q) = %q, want %q", in, got, want)
		}
	}
}
