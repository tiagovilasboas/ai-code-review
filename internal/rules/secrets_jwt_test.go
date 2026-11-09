package rules

import "testing"

func TestSecretsJWTSigning(t *testing.T) {
	runSecretShapes(t, []secretShape{
		{"sign literal", "src/token.ts", `+  return jwt.sign(payload, "Rt8Yu1Zc4Bn6Hd2Ws9Kp", { expiresIn: '1h' });`, 1},
		{"verify literal", "src/token.js", `+  jwt.verify(token, 'Rt8Yu1Zc4Bn6Hd2Ws9Kp');`, 1},
		{"const jwt secret", "src/token.ts", `+const JWT_SECRET = "k3Jx9Qm2Lp7Vw5Rt8Yu1";`, 1},
		{"camel jwt key", "src/token.ts", `+const jwtKey = "k3Jx9Qm2Lp7Vw5Rt8Yu1";`, 1},
		{"text encoder key", "src/token.ts", `+export const key = new TextEncoder().encode("Zq81xPmL3vT7nB5cD9fH");`, 1},
		{"hmac literal", "src/sig.ts", `+  const h = createHmac('sha256', 'Zq81xPmL3vT7nB5cD9fH');`, 1},
		{"sign variable", "src/token.ts", `+  return jwt.sign(payload, secret, { expiresIn: '1h' });`, 0},
		{"sign env", "src/token.ts", `+  return jwt.sign(payload, process.env.JWT_SECRET!);`, 0},
		{"encoder config", "src/token.ts", `+export const key = new TextEncoder().encode(config.jwtSecret);`, 0},
		{"encoder plain text", "src/util.ts", `+const bytes = new TextEncoder().encode("hello");`, 0},
		{"encoder prose", "src/util.ts", `+const key = new TextEncoder().encode("hello world");`, 0},
		{"hmac variable", "src/sig.ts", `+  const h = createHmac('sha256', key);`, 0},
		{"jwt test file", "src/token.test.ts", `+  jwt.sign(payload, "Rt8Yu1Zc4Bn6Hd2Ws9Kp");`, 0},
	})
}
