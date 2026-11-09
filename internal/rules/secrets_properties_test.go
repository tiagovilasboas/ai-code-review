package rules

import "testing"

func TestSecretsConfigFiles(t *testing.T) {
	runSecretShapes(t, []secretShape{
		{"properties password", "src/main/resources/application.properties", `+spring.datasource.password=Zq81xPmL3vT7nB5`, 1},
		{"properties api-key", "src/main/resources/application.properties", `+payments.client.api-key=k3Jx9Qm2Lp7Vw5`, 1},
		{"properties secret", "src/main/resources/application-prod.properties", `+jwt.secret=Rt8Yu1Zc4Bn6Hd2`, 1},
		{"yaml unquoted password", "src/main/resources/application.yml", `+    password: Zq81xPmL3vT7nB5`, 1},
		{"yaml quoted secret", "src/main/resources/application-prod.yaml", `+    secret: 'Rt8Yu1Zc4Bn6Hd2Ws9Kp'`, 1},
		{"properties env placeholder", "application.properties", `+spring.datasource.password=${DB_PASSWORD}`, 0},
		{"yaml env placeholder", "application.yml", `+    password: ${DB_PASSWORD:}`, 0},
		{"properties empty", "application.properties", `+payments.client.secret=`, 0},
		{"properties dummy", "application.properties", `+spring.security.user.password=changeme`, 0},
		{"properties encrypted", "application.properties", `+spring.datasource.password=ENC(k3Jx9Qm2Lp7Vw5)`, 0},
		{"yaml secret reference name", "values.yaml", `+  existingSecret: db-credentials`, 0},
		{"properties unrelated key", "application.properties", `+management.endpoint.token-endpoint.enabled=true`, 0},
		{"properties comment", "application.properties", `+# spring.datasource.password=Zq81xPmL3vT7nB5`, 0},
		{"test resources", "src/test/resources/application.properties", `+spring.datasource.password=Zq81xPmL3vT7nB5`, 0},
		{"code file qualified key stays silent", "internal/a.go", `+	cfg.Secret: "label-value",`, 0},
	})
}
