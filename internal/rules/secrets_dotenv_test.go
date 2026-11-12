package rules

import "testing"

const sendGridSample = "SG.abcdefghij0123456789AB.abcdefghij0123456789ab"

func TestSecretsDotenvAndSendGrid(t *testing.T) {
	runSecretShapes(t, []secretShape{
		{"sendgrid key in .env.production", ".env.production", `+SENDGRID_API_KEY=` + sendGridSample, 1},
		{"sendgrid key, any env var name", ".env", `+MAIL_PROVIDER_KEY=` + sendGridSample, 1},
		{"sendgrid key quoted in code", "config/mailer.js", `+const key = "` + sendGridSample + `";`, 1},
		{"sendgrid shaped placeholder", "config/mailer.js", `+const key = "SG.xxxxxxxxxxxxxxxxxxxxxx.xxxxxxxxxxxxxxxxxxxxxxxxxxx";`, 0},
		{"sendgrid prefix without the shape", ".env", `+MAIL_PROVIDER=SG.short.tail`, 0},
		{"dotenv credential literal", ".env.production", `+JWT_SECRET=8f3a91c2b7d44e5fa0a1`, 1},
		{"dotenv export prefix", ".env", `+export API_TOKEN=Zq81xPmL3vT7nB5Rt8Y`, 1},
		{"dotenv stripe live key", ".env.local", `+PAYMENTS=sk_live_abcdefghijklmnop1234`, 1},
		{"dotenv variable reference", ".env.production", `+DB_PASSWORD=${DB_PASSWORD}`, 0},
		{"dotenv empty value", ".env.production", `+API_TOKEN=`, 0},
		{"dotenv weak dev value", ".env", `+DB_PASSWORD=postgres`, 0},
		{"dotenv template", ".env.example", `+API_TOKEN=Zq81xPmL3vT7nB5Rt8Y`, 0},
		{"dotenv non secret", ".env.production", `+MAIL_FROM=noreply@example.com`, 0},
		{"dotenv comment", ".env.production", `+# API_TOKEN=Zq81xPmL3vT7nB5Rt8Y`, 0},
		{"plain text file stays silent", "notes.txt", `+API_TOKEN=Zq81xPmL3vT7nB5Rt8Y`, 0},
	})
}
