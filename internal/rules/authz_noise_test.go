package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestAuthZGoClientsAndConvertedIDs(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		{"rest client builder chain", "api/pods.go", []string{
			"+func (h *H) Logs(w http.ResponseWriter, r *http.Request) {",
			`+	ns := chi.URLParam(r, "namespace")`,
			`+	name := chi.URLParam(r, "name")`,
			`+	req := h.client.Get().Namespace(ns).Name(name).SubResource("log")`,
			"+	res := req.Do(r.Context())",
			"+}"}, 0},
		{"rest client builder over several lines", "api/pods.go", []string{
			"+func (h *H) List(c *gin.Context) {",
			`+	ns := c.Param("namespace")`,
			"+	res := h.client.Get().",
			"+		Namespace(ns).",
			`+		Resource("pods").`,
			"+		Do(c.Request.Context())",
			"+}"}, 0},
		{"bearer token forwarded upstream", "api/pods.go", []string{
			"+func (h *H) Pod(w http.ResponseWriter, r *http.Request) {",
			`+	name := chi.URLParam(r, "name")`,
			"+	cfg := rest.Config{BearerToken: tokenFrom(r)}",
			"+	cs, _ := kubernetes.NewForConfig(&cfg)",
			`+	pod, _ := cs.CoreV1().Pods("x").Get(r.Context(), name, metav1.GetOptions{})`,
			"+}"}, 0},
		{"repository Get with route name still fires", "api/pods.go", []string{
			"+func (h *H) Pod(w http.ResponseWriter, r *http.Request) {",
			`+	name := chi.URLParam(r, "name")`,
			"+	pod, _ := h.repo.Get(r.Context(), name)",
			"+}"}, 1},
		{"chi id parsed before the lookup", "http/invoices.go", []string{
			"+func (h *H) Get(w http.ResponseWriter, r *http.Request) {",
			`+	idStr := chi.URLParam(r, "id")`,
			"+	id, err := strconv.ParseInt(idStr, 10, 64)",
			"+	if err != nil {",
			"+		return",
			"+	}",
			"+	inv, err := h.repo.FindByID(r.Context(), id)",
			"+}"}, 1},
		{"chi id parsed to uuid", "http/invoices.go", []string{
			"+func (h *H) Get(w http.ResponseWriter, r *http.Request) {",
			`+	id, err := uuid.Parse(chi.URLParam(r, "id"))`,
			"+	inv, err := h.repo.FindByID(r.Context(), id)",
			"+}"}, 1},
		{"chi parsed id with owner scoped lookup", "http/invoices.go", []string{
			"+func (h *H) Get(w http.ResponseWriter, r *http.Request) {",
			`+	idStr := chi.URLParam(r, "id")`,
			"+	id, _ := strconv.ParseInt(idStr, 10, 64)",
			"+	user := auth.UserFrom(r.Context())",
			"+	inv, err := h.repo.FindByIDAndOwner(r.Context(), id, user.ID)",
			"+}"}, 0},
	}
	runFrameworkCases(t, cases)
}

func TestAuthZSpringRoleOnlyAnnotations(t *testing.T) {
	get := func(ann string) []string {
		return []string{
			"+@RestController",
			"+public class InvoiceController {",
			"+    " + ann,
			`+    @GetMapping("/{id}")`,
			`+    public InvoiceDto get(@PathVariable("id") Long id) throws NotFoundException {`,
			"+        return invoiceRepository.findById(id).orElseThrow();",
			"+    }",
			"+    @DeleteMapping(\"/{id}\")",
			"+    public void remove(@PathVariable Long id, Principal principal) {",
			"+        invoiceRepository.findByIdAndOwnerId(id, principal.getName());",
			"+    }",
			"+}"}
	}
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		{"user role only", "InvoiceController.java", get(`@PreAuthorize("hasRole('USER')")`), 1},
		{"is authenticated only", "InvoiceController.java", get(`@PreAuthorize("isAuthenticated()")`), 1},
		{"admin role", "InvoiceController.java", get(`@PreAuthorize("hasRole('ADMIN')")`), 0},
		{"ownership expression", "InvoiceController.java", get(`@PreAuthorize("@invoiceSecurity.isOwner(#id, authentication)")`), 0},
		{"no annotation, handler with throws clause", "InvoiceController.java", get(`@Transactional`), 1},
	}
	runFrameworkCases(t, cases)
}

func TestAuthZSpringOwnerScopedNeighbourDoesNotHideLookup(t *testing.T) {
	runFrameworkCases(t, []struct {
		name, path string
		body       []string
		want       int
	}{
		{"owner check in the next method only", "A.java", []string{
			`+    @GetMapping("/{id}")`,
			`+    public Invoice get(@PathVariable Long id) throws IOException {`,
			"+        return invoiceRepository.findById(id).orElseThrow();",
			"+    }",
			`+    @GetMapping("/mine/{id}")`,
			`+    public Invoice mine(@PathVariable Long id, Principal principal) {`,
			"+        return invoiceRepository.findByIdAndOwnerId(id, principal.getName()).orElseThrow();",
			"+    }"}, 1},
	})
}

func runFrameworkCases(t *testing.T, cases []struct {
	name, path string
	body       []string
	want       int
}) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countClass(AuthZ(fileOf("src/"+c.path, 1, c.body...)), finding.AuthZ)
			if got != c.want {
				t.Fatalf("got %d authz findings, want %d", got, c.want)
			}
		})
	}
}
