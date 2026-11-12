package rules

import "testing"

func TestSecretsPHPConfigArrays(t *testing.T) {
	const cfg = "config/autoload/mail.global.php"
	runSecretShapes(t, []secretShape{
		{"smtp password literal", cfg, `+        'password' => 'Zq7!vN3pXk82mWc4Rt',`, 1},
		{"double quoted key and value", cfg, `+        "smtp_password" => "Zq7!vN3pXk82mWc4Rt",`, 1},
		{"api key", cfg, `+    'api_key' => 'k9XvQ2mLr7TwB4nZ',`, 1},
		{"client secret", cfg, `+    'client_secret' => 'a8F3kP0zQw9RtY6u',`, 1},
		{"getenv", cfg, `+        'password' => getenv('SMTP_PASSWORD'),`, 0},
		{"env helper", cfg, `+        'password' => env('SMTP_PASSWORD', ''),`, 0},
		{"env helper with literal default is not matched", cfg, `+        'password' => env('SMTP_PASSWORD'),`, 0},
		{"empty", cfg, `+        'password' => '',`, 0},
		{"changeme", cfg, `+        'password' => 'changeme',`, 0},
		{"weak word", cfg, `+        'password' => 'secret',`, 0},
		{"angle placeholder", cfg, `+        'password' => '<your-password>',`, 0},
		{"translation key", cfg, `+        'password' => 'validation.password',`, 0},
		{"label without digit", cfg, `+        'password' => 'Password',`, 0},
		{"prose label", cfg, `+        'password' => 'The provided password is incorrect.',`, 0},
		{"bcrypt hash placeholder var", cfg, `+        'password' => '$2y$10$abcdefghijklmnopqrstuv',`, 0},
		{"key that does not end in a credential noun", cfg, `+        'password_min_length' => 'Zq7!vN3pXk82mWc4Rt',`, 0},
		{"display mask", cfg, `+        'password' => '**********',`, 0},
		{"validation rule", cfg, `+            'password' => 'required|string|max:1000',`, 0},
		{"validation rule with parameter", cfg, `+            'password' => 'required_without:passkey_challenge_token|string|max:1000',`, 0},
		{"not a php file", "config/mail.json", `+        'password' => 'Zq7!vN3pXk82mWc4Rt',`, 0},
		{"test path", "tests/config/mail.php", `+        'password' => 'Zq7!vN3pXk82mWc4Rt',`, 0},
	})
}
