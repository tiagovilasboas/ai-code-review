package rules

import "testing"

func TestAuthZGorillaMux(t *testing.T) {
	const gof = "internal/orders/handler.go"
	runFwCases(t, AuthZ, []fwCase{
		{"mux vars id, caller only logged", gof, []string{
			` func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {`,
			` 	vars := mux.Vars(r)`,
			` 	log.Printf("order lookup by %s", auth.UserFrom(r.Context()).ID)`,
			`+	order, err := h.store.GetOrder(r.Context(), vars["id"])`}, []int{4}},
		{"mux vars bound to a variable", gof, []string{
			` func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {`,
			`+	id := mux.Vars(r)["id"]`,
			`+	order, err := h.store.GetOrder(r.Context(), id)`}, []int{3}},
		{"mux inline index", gof, []string{
			` func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {`,
			`+	err := h.store.DeleteOrder(r.Context(), mux.Vars(r)["id"])`}, []int{2}},
		{"logger and slog and fmt are not guards", gof, []string{
			` func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {`,
			`+	h.logger.Info("by", "user", auth.UserFrom(r.Context()).ID)`,
			`+	slog.Info("by", "user", claims.Subject)`,
			`+	fmt.Println(auth.UserFrom(r.Context()))`,
			`+	order, err := h.store.GetOrder(r.Context(), mux.Vars(r)["id"])`}, []int{5}},
		{"caller id passed to the store call", gof, []string{
			` func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {`,
			` 	vars := mux.Vars(r)`,
			`+	order, err := h.store.GetOrder(r.Context(), vars["id"], auth.UserFrom(r.Context()).ID)`}, nil},
		{"ownership compare after the lookup", gof, []string{
			` func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {`,
			`+	order, err := h.store.GetOrder(r.Context(), mux.Vars(r)["id"])`,
			`+	if order.OwnerID != auth.UserFrom(r.Context()).ID {`,
			`+		return`,
			`+	}`}, nil},
		{"a userID route param is not a principal", gof, []string{
			` func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {`,
			`+	uid := mux.Vars(r)["userID"]`,
			`+	order, err := h.store.GetOrder(r.Context(), uid)`}, []int{3}},
		{"mux vars slug", gof, []string{
			` func (h *Handler) GetPage(w http.ResponseWriter, r *http.Request) {`,
			`+	page, err := h.store.GetPage(r.Context(), mux.Vars(r)["slug"])`}, nil},
	})
}

func TestAuthZSpringNotFoundIsNotAGuard(t *testing.T) {
	const jv = "src/main/java/com/acme/billing/InvoiceController.java"
	runFwCases(t, AuthZ, []fwCase{
		{"orElseThrow not found", jv, []string{
			` @GetMapping("/{id}")`,
			` public InvoiceDto get(@PathVariable Long id) {`,
			`+    Invoice invoice = invoiceRepository.findById(id).orElseThrow(NotFoundException::new);`,
			`+    return InvoiceDto.from(invoice);`,
			` }`}, []int{3}},
		{"ResponseStatusException 404", jv, []string{
			` @GetMapping("/{id}")`,
			` public InvoiceDto get(@PathVariable Long id) {`,
			`+    Invoice invoice = invoiceRepository.findById(id)`,
			`+        .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));`,
			` }`}, []int{3}},
		{"ResponseStatusException 403", jv, []string{
			` @GetMapping("/{id}")`,
			` public InvoiceDto get(@PathVariable Long id) {`,
			`+    Invoice invoice = invoiceRepository.findById(id).orElseThrow();`,
			`+    if (!mine(invoice)) throw new ResponseStatusException(HttpStatus.FORBIDDEN);`,
			` }`}, nil},
		{"ResponseStatusException 401", jv, []string{
			` @GetMapping("/{id}")`,
			` public InvoiceDto get(@PathVariable Long id) {`,
			`+    Invoice invoice = invoiceRepository.findById(id).orElseThrow();`,
			`+    if (!mine(invoice)) throw new ResponseStatusException(HttpStatus.UNAUTHORIZED);`,
			` }`}, nil},
		{"owner scoped repository call", jv, []string{
			` @GetMapping("/{id}")`,
			` public InvoiceDto get(@PathVariable Long id, Authentication auth) {`,
			`+    return invoiceRepository.findByIdAndOwnerId(id, auth.getName()).map(InvoiceDto::from).orElseThrow();`,
			` }`}, nil},
	})
}
