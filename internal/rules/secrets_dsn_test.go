package rules

import "testing"

func TestSecretsDSN(t *testing.T) {
	runSecretShapes(t, []secretShape{
		{"url literal", "internal/db.go", `+const u = "postgres://billing:Zq81xPmL3vT7@db.internal:5432/billing"`, 1},
		{"kv dsn literal", "internal/db.go", `+	return "host=db.internal user=billing password=Zq81xPmL3vT7 dbname=billing"`, 1},
		{"sqlserver style", "internal/db.cs", `+var c = "Server=db.internal;Database=app;User Id=app;Password=Zq81xPmL3vT7;"`, 1},
		{"mysql tcp dsn", "internal/db.go", `+	dsn := "app:Zq81xPmL3vT7@tcp(db.internal:3306)/billing"`, 1},
		{"url without password", "internal/db.go", `+const u = "postgres://billing@db.internal:5432/billing"`, 0},
		{"kv env placeholder", "internal/db.go", `+	return "host=db.internal user=billing password=${DB_PASSWORD} dbname=billing"`, 0},
		{"kv printf verb", "internal/db.go", `+	return fmt.Sprintf("host=db user=app password=%s dbname=app", pw)`, 0},
		{"kv local dev", "internal/db.go", `+	return "host=localhost user=app password=Zq81xPmL3vT7 dbname=app"`, 0},
		{"url env built", "internal/db.go", `+	return fmt.Sprintf("postgres://billing:%s@db.internal/billing", os.Getenv("DB_PASSWORD"))`, 0},
		{"tcp dsn env", "internal/db.go", `+	dsn := user + ":" + os.Getenv("DB_PASSWORD") + "@tcp(db:3306)/app"`, 0},
		{"kv test file", "internal/db_test.go", `+	d := "host=db user=app password=Zq81xPmL3vT7 dbname=app"`, 0},
	})
}
