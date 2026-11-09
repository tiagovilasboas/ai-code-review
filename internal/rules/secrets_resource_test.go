package rules

import "testing"

func TestSecretsSecretManagerResourceNames(t *testing.T) {
	runSecretShapes(t, []secretShape{
		{"gcp version path", "internal/gcp.go", `+var dbPasswordRef = "projects/my-proj/secrets/db-password/versions/latest"`, 0},
		{"gcp secret const", "internal/gcp.go", `+const apiSecret = "projects/123456/secrets/api-secret/versions/3"`, 0},
		{"gcp password var", "internal/gcp.go", `+	password := "projects/my-proj/locations/eu/secrets/db/versions/1"`, 0},
		{"aws arn", "internal/aws.go", `+const dbSecret = "arn:aws:secretsmanager:eu-west-1:123456789012:secret:prod/db-AbCdEf"`, 0},
		{"slashes in real secret", "internal/aws.go", `+const secret = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYzq81xPm"`, 1},
	})
}
