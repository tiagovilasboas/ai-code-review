package rules

import "testing"

func TestAuthZSpringLaravelDjangoFastAPIChi(t *testing.T) {
	cases := []struct {
		name, path string
		body       []string
		want       int
	}{
		{"spring findById by path variable", "OrderController.java", []string{
			`+@GetMapping("/orders/{id}")`,
			`+public Order get(@PathVariable Long id) {`,
			`+  return orderRepository.findById(id).orElseThrow();`,
			`+}`}, 1},
		{"spring deleteById", "OrderController.java", []string{
			`+@DeleteMapping("/orders/{id}")`,
			`+public void del(@PathVariable("id") Long id) {`,
			`+  orderRepository.deleteById(id);`,
			`+}`}, 1},
		{"spring request param id", "OrderController.java", []string{
			`+@GetMapping("/orders")`,
			`+public Order get(@RequestParam Long orderId) {`,
			`+  return svc.findById(orderId);`,
			`+}`}, 1},
		{"spring authentication principal", "OrderController.java", []string{
			`+@GetMapping("/orders/{id}")`,
			`+public Order get(@PathVariable Long id, @AuthenticationPrincipal UserDetails me) {`,
			`+  return orderRepository.findByIdAndOwner(id, me.getUsername()).orElseThrow();`,
			`+}`}, 0},
		{"spring PreAuthorize above", "OrderController.java", []string{
			`+@PreAuthorize("@orderPolicy.canView(#id)")`,
			`+@GetMapping("/orders/{id}")`,
			`+public Order get(@PathVariable Long id) {`,
			`+  return orderRepository.findById(id).orElseThrow();`,
			`+}`}, 0},
		{"spring class-level PreAuthorize", "OrderController.java", []string{
			`+@PreAuthorize("hasRole('ADMIN')")`,
			`+public class AdminOrders {`,
			`+  @GetMapping("/orders/{id}")`,
			`+  public Order get(@PathVariable Long id) {`,
			`+    return orderRepository.findById(id).orElseThrow();`,
			`+  }`,
			`+}`}, 0},
		{"spring scoped repository method", "OrderController.java", []string{
			`+@GetMapping("/orders/{id}")`,
			`+public Order get(@PathVariable Long id) {`,
			`+  return orderRepository.findByIdAndUserId(id, currentUser.getId());`,
			`+}`}, 0},
		{"spring slug", "PageController.java", []string{
			`+@GetMapping("/pages/{slug}")`,
			`+public Page get(@PathVariable String slug) {`,
			`+  return pageRepository.findBySlug(slug);`,
			`+}`}, 0},
		{"spring id never reaches a lookup", "OrderController.java", []string{
			`+@GetMapping("/echo/{id}")`,
			`+public String get(@PathVariable Long id) {`,
			`+  return "x" + id;`,
			`+}`}, 0},
		{"laravel find by action id", "app/Http/Controllers/OrderController.php", []string{
			`+public function show($id)`,
			`+{`,
			`+    return Order::findOrFail($id);`,
			`+}`}, 1},
		{"laravel request input id", "app/Http/Controllers/OrderController.php", []string{
			`+public function destroy(Request $request)`,
			`+{`,
			`+    $oid = $request->input('order_id');`,
			`+    Order::destroy($oid);`,
			`+}`}, 1},
		{"laravel route closure", "routes/web.php", []string{
			`+Route::get('/orders/{id}', function ($id) {`,
			`+    return Order::find($id);`,
			`+});`}, 1},
		{"laravel authorize", "app/Http/Controllers/OrderController.php", []string{
			`+public function show($id)`,
			`+{`,
			`+    $order = Order::findOrFail($id);`,
			`+    $this->authorize('view', $order);`,
			`+    return $order;`,
			`+}`}, 0},
		{"laravel scoped through user", "app/Http/Controllers/OrderController.php", []string{
			`+public function show(Request $request, $id)`,
			`+{`,
			`+    return $request->user()->orders()->findOrFail($id);`,
			`+}`}, 0},
		{"laravel gate", "app/Http/Controllers/OrderController.php", []string{
			`+public function show($id)`,
			`+{`,
			`+    abort_unless(Gate::allows('view-order', $id), 403);`,
			`+    return Order::find($id);`,
			`+}`}, 0},
		{"laravel private helper is not an action", "app/Http/Controllers/OrderController.php", []string{
			`+private function load($id)`,
			`+{`,
			`+    return Order::find($id);`,
			`+}`}, 0},
		{"laravel where on a non id column", "app/Http/Controllers/OrderController.php", []string{
			`+public function index(Request $request)`,
			`+{`,
			`+    return Order::where('status', $request->input('status'))->get();`,
			`+}`}, 0},
		{"django get by pk", "orders/views.py", []string{
			`+def order_detail(request, pk):`,
			`+    order = Order.objects.get(pk=pk)`,
			`+    return JsonResponse(model_to_dict(order))`}, 1},
		{"django get_object_or_404", "orders/views.py", []string{
			`+def order_detail(request, order_id):`,
			`+    order = get_object_or_404(Order, id=order_id)`,
			`+    return render(request, "o.html", {"o": order})`}, 1},
		{"django filtered by request.user", "orders/views.py", []string{
			`+def order_detail(request, pk):`,
			`+    order = get_object_or_404(Order, pk=pk, owner=request.user)`,
			`+    return render(request, "o.html", {"o": order})`}, 0},
		{"fastapi path id", "app/routes.py", []string{
			`+@router.get("/items/{item_id}")`,
			`+def read_item(item_id: int, db: Session = Depends(get_db)):`,
			`+    return db.query(Item).filter(Item.id == item_id).first()`}, 1},
		{"fastapi current user", "app/routes.py", []string{
			`+@router.get("/items/{item_id}")`,
			`+def read_item(item_id: int, user: User = Depends(get_current_user), db: Session = Depends(get_db)):`,
			`+    return db.query(Item).filter(Item.id == item_id, Item.owner_id == user.id).first()`}, 0},
		{"fastapi slug", "app/routes.py", []string{
			`+@router.get("/pages/{slug}")`,
			`+def page(slug: str, db: Session = Depends(get_db)):`,
			`+    return db.query(Page).filter(Page.slug == slug).first()`}, 0},
		{"chi URLParam", "h.go", []string{
			`+func Get(w http.ResponseWriter, r *http.Request) {`,
			`+	id := chi.URLParam(r, "id")`,
			`+	a, _ := repo.GetAccount(id)`,
			`+	json.NewEncoder(w).Encode(a)`,
			`+}`}, 1},
		{"net/http PathValue", "h.go", []string{
			`+func Get(w http.ResponseWriter, r *http.Request) {`,
			`+	id := r.PathValue("id")`,
			`+	db.First(&acct, id)`,
			`+}`}, 1},
		{"chi with context principal", "h.go", []string{
			`+func Get(w http.ResponseWriter, r *http.Request) {`,
			`+	id := chi.URLParam(r, "id")`,
			`+	uid := r.Context().Value(userKey)`,
			`+	a, _ := repo.GetForUser(id, uid)`,
			`+}`}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := len(AuthZ(mkFile(c.path, c.body...))); got != c.want {
				t.Fatalf("findings = %d, want %d", got, c.want)
			}
		})
	}
}
