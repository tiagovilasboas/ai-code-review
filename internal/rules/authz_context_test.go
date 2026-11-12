package rules

import "testing"

// The handler and route are unchanged context lines; only the lookup is added.
func TestAuthZContextHandlers(t *testing.T) {
	runFrameworkCases(t, []struct {
		name, path string
		body       []string
		want       int
	}{
		{"go chi ByID method", "h.go", []string{
			" func (h *H) Get(w http.ResponseWriter, r *http.Request) {",
			`     id := chi.URLParam(r, "id")`,
			"+    inv, err := h.store.InvoiceByID(r.Context(), id)",
			" }"}, 1},
		{"go guard in context", "h.go", []string{
			" func (h *H) Get(w http.ResponseWriter, r *http.Request) {",
			`     id := chi.URLParam(r, "id")`,
			"     uid := claims.UserID",
			"+    inv, err := h.store.InvoiceByID(r.Context(), id)",
			" }"}, 0},
		{"go added line reads a non-id param", "h.go", []string{
			" func (h *H) List(w http.ResponseWriter, r *http.Request) {",
			`     page := chi.URLParam(r, "page")`,
			"+    rows, err := h.store.ListPage(r.Context(), page)",
			" }"}, 0},
		{"go added line does not use the id", "h.go", []string{
			" func (h *H) Get(w http.ResponseWriter, r *http.Request) {",
			`     id := chi.URLParam(r, "id")`,
			"+    log.Println(\"get\")",
			" }"}, 0},
		{"go handler opener not visible", "h.go", []string{
			`     id := chi.URLParam(r, "id")`,
			"+    inv, err := h.store.InvoiceByID(r.Context(), id)"}, 1},
		{"flask context route", "v.py", []string{
			` @bp.route("/invoices/<int:invoice_id>")`,
			" @login_required",
			" def get_invoice(invoice_id):",
			"+    invoice = Invoice.query.get_or_404(invoice_id)"}, 1},
		{"flask guard in context", "v.py", []string{
			` @bp.route("/invoices/<int:invoice_id>")`,
			" def get_invoice(invoice_id):",
			"     check(current_user)",
			"+    invoice = Invoice.query.get_or_404(invoice_id)"}, 0},
		{"flask route without id", "v.py", []string{
			` @bp.route("/invoices")`,
			" def list_invoices():",
			"+    rows = Invoice.query.filter_by(status='open')"}, 0},
		{"flask route not visible", "v.py", []string{
			" def get_invoice(invoice_id):",
			"+    invoice = Invoice.query.get_or_404(invoice_id)"}, 0},
		{"spring context mapping", "A.java", []string{
			`     @GetMapping("/invoices/{id}")`,
			"     public ResponseEntity<Invoice> get(@PathVariable Long id) {",
			"+        return invoiceRepository.findById(id)",
			"     }"}, 1},
		{"spring principal in context", "A.java", []string{
			`     @GetMapping("/invoices/{id}")`,
			"     public ResponseEntity<Invoice> get(@PathVariable Long id) {",
			"         var a = SecurityContextHolder.getContext().getAuthentication();",
			"+        return invoiceRepository.findById(id)",
			"     }"}, 0},
		{"spring request param that is not an id", "A.java", []string{
			`     @GetMapping("/invoices")`,
			"     public List<Invoice> list(@RequestParam String status) {",
			"+        return invoiceRepository.findByStatus(status);"}, 0},
		{"laravel context action", "C.php", []string{
			"     public function show($id)",
			"     {",
			"+        $invoice = Invoice::findOrFail($id);",
			"     }"}, 1},
		{"laravel authorize in context", "C.php", []string{
			"     public function show($id)",
			"     {",
			"         $this->authorize('view');",
			"+        $invoice = Invoice::findOrFail($id);"}, 0},
		{"laravel added line ignores the id", "C.php", []string{
			"     public function show($id)",
			"     {",
			"+        $all = Invoice::where('open', 1)->get();"}, 0},
	})
}
