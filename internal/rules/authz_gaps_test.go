package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestAuthZJaxrsFastapiEchoGaps(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		{"jaxrs delete by id", "R.java", []string{
			"+    @DELETE", `+    @Path("/{id}")`,
			`+    public Response remove(@PathParam("id") Long id) {`,
			"+        repository.deleteById(id);", "+        return Response.noContent().build();", "+    }"}, 1},
		{"jaxrs em.remove find", "R.java", []string{
			"+    @DELETE", `+    @Path("/{id}")`,
			`+    public Response remove(@PathParam("id") Long id) {`,
			"+        em.remove(em.find(Document.class, id));", "+        return Response.ok().build();", "+    }"}, 1},
		{"jaxrs roles allowed", "R.java", []string{
			"+    @DELETE", `+    @Path("/{id}")`, `+    @RolesAllowed("admin")`,
			`+    public Response remove(@PathParam("id") Long id) {`,
			"+        repository.deleteById(id);", "+    }"}, 0},
		{"jaxrs roles allowed on class", "R.java", []string{
			`+@RolesAllowed("admin")`, "+public class R {",
			"+    @DELETE", `+    @Path("/{id}")`,
			`+    public Response remove(@PathParam("id") Long id) {`,
			"+        repository.deleteById(id);", "+    }", "+}"}, 0},
		{"jaxrs security context param", "R.java", []string{
			"+    @DELETE", `+    @Path("/{id}")`,
			`+    public void remove(@PathParam("id") Long id, @Context SecurityContext sc) {`,
			"+        repository.deleteById(id);", "+    }"}, 0},
		{"jaxrs owner filter", "R.java", []string{
			"+    @DELETE", `+    @Path("/{id}")`,
			`+    public void remove(@PathParam("id") Long id, @Context SecurityContext sc) {`,
			"+        repository.deleteByIdAndOwnerId(id, owner);", "+    }"}, 0},
		{"fastapi multi-line query chain", "a.py", []string{
			`+@router.delete("/documents/{doc_id}")`,
			"+def delete_document(doc_id: int, db: Session = Depends(get_db)):",
			"+    (", "+        db.query(Document)", "+        .filter(Document.id == doc_id)", "+        .delete()", "+    )"}, 1},
		{"fastapi session.exec select", "a.py", []string{
			`+@router.delete("/documents/{doc_id}")`,
			"+def delete_document(doc_id: int, session: Session = Depends(get_session)):",
			"+    doc = session.exec(select(Document).where(Document.id == doc_id)).first()",
			"+    session.delete(doc)"}, 1},
		{"fastapi delete with current user", "a.py", []string{
			`+@router.delete("/documents/{doc_id}")`,
			"+def delete_document(doc_id: int, user=Depends(get_current_user), db: Session = Depends(get_db)):",
			"+    db.query(Document).filter(Document.id == doc_id, Document.owner_id == user.id).delete()"}, 0},
		{"echo file name through Base", "h.go", []string{
			"+func Download(c echo.Context) error {",
			`+	name := filepath.Base(c.Param("name"))`,
			`+	data, err := os.ReadFile(filepath.Join(exportDir, name))`,
			"+	return c.Blob(200, \"text/csv\", data)", "+}"}, 0},
		{"echo file name Base inline", "h.go", []string{
			"+func Download(c echo.Context) error {",
			`+	name := c.Param("name")`,
			`+	data, err := os.ReadFile(filepath.Join(exportDir, filepath.Base(name)))`,
			"+	return c.Blob(200, \"text/csv\", data)", "+}"}, 0},
		{"echo id still fires on repo read", "h.go", []string{
			"+func Get(c echo.Context) error {",
			`+	id := c.Param("id")`,
			"+	a, _ := repo.GetAccount(id)",
			"+	return c.JSON(200, a)", "+}"}, 1},
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
