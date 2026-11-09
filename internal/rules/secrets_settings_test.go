package rules

import "testing"

func TestSecretsPythonSettings(t *testing.T) {
	runSecretShapes(t, []secretShape{
		{"django secret key", "app/settings.py", `+SECRET_KEY = 'k3Jx9Qm2Lp7Vw5Rt8Yu1'`, 1},
		{"aws secret literal", "app/settings.py", `+AWS_SECRET_ACCESS_KEY = "Zq81xPmL3vT7nB5cD9fH"`, 1},
		{"databases password", "app/settings.py", `+        'PASSWORD': 'Tr0ub4dor83xQ',`, 1},
		{"db password const", "app/settings.py", `+DB_PASSWORD = "Tr0ub4dor83xQ"`, 1},
		{"environ lookup", "app/settings.py", `+SECRET_KEY = os.environ['DJANGO_SECRET_KEY']`, 0},
		{"getenv with empty default", "app/settings.py", `+AWS_SECRET_ACCESS_KEY = os.getenv('AWS_SECRET_ACCESS_KEY', '')`, 0},
		{"env dict value", "app/settings.py", `+        'PASSWORD': os.environ['DB_PASSWORD'],`, 0},
		{"empty value", "app/settings.py", `+SECRET_KEY = ''`, 0},
		{"placeholder", "app/settings.py", `+SECRET_KEY = 'changeme'`, 0},
		{"non credential name", "app/settings.py", `+CACHE_KEY = 'tenant-cache'`, 0},
		{"cookie name setting", "app/settings.py", `+CSRF_TOKEN = 'csrf-token'`, 0},
		{"test path", "app/tests/settings.py", `+SECRET_KEY = 'k3Jx9Qm2Lp7Vw5Rt8Yu1'`, 0},
	})
}
