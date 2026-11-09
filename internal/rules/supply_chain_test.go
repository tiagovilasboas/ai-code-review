package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestSupplyChainPackageJSONPostinstall(t *testing.T) {
	f := mkFile("package.json",
		` {`, `   "name": "sample-app",`, `   "version": "1.0.0",`, `   "private": true,`,
		`   "scripts": {`, `     "start": "node src/server.js",`,
		`+    "test": "node --test",`,
		`+    "postinstall": "node scripts/fetch-tool.js"`,
		`   },`,
	)
	got := SupplyChain(f)
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %+v", got)
	}
	wantFinding(t, got[0], "Install-time script", finding.InstallScript, 8, "CWE-829", finding.High, finding.Block)
}

func TestSupplyChainNpmrcRegistry(t *testing.T) {
	f := mkFile(".npmrc",
		`+# Unexpected registry host (synthetic fixture)`,
		`+registry=https://pkgs.example.invalid/npm/`,
	)
	got := SupplyChain(f)
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %+v", got)
	}
	wantFinding(t, got[0], "Unexpected npm registry host", finding.RegistryHost, 2, "CWE-829", finding.High, finding.Block)

	ok := mkFile(".npmrc", `+registry=https://registry.npmjs.org/`, `+@acme:registry=https://registry.npmjs.org/`)
	if got := SupplyChain(ok); len(got) != 0 {
		t.Fatalf("npmjs registry must be silent, got %+v", got)
	}
	scoped := mkFile(".npmrc", `+@acme:registry=https://npm.acme.invalid/`)
	if got := SupplyChain(scoped); len(got) != 1 {
		t.Fatalf("scoped foreign registry must be flagged, got %+v", got)
	}
}

func TestSupplyChainWorkflow(t *testing.T) {
	f := mkFile(".github/workflows/ci.yml",
		` name: ci`, ` on:`, `   pull_request:`, ` jobs:`, `   test:`,
		`     runs-on: ubuntu-latest`, `     steps:`,
		`       - uses: actions/checkout@v4`,
		`+      - uses: some-org/cache-warmup@v1`,
		`       - uses: actions/setup-node@v4`,
		`         with:`, `           node-version: "20"`, `           cache: npm`,
		`+      - run: npm install`,
		`       - run: npm test`,
	)
	got := SupplyChain(f)
	if len(got) != 2 {
		t.Fatalf("want 2 findings, got %+v", got)
	}
	wantFinding(t, got[0], "Unpinned third-party GitHub Action", finding.UnpinnedAction, 9, "CWE-829", finding.High, finding.Block)
	wantFinding(t, got[1], "CI install can rewrite the lockfile", finding.LockfileRewrite, 14, "CWE-1104", finding.Medium, finding.Review)
}

func TestSupplyChainWorkflowPassCases(t *testing.T) {
	sha := "+      - uses: some-org/x@0123456789abcdef0123456789ABCDEF01234567"
	cases := []string{
		sha, "+      - uses: github/codeql-action@v3", "+      - uses: actions/cache@v4",
		"+      - run: npm install --frozen-lockfile", "+      # uses: evil/action@v1",
		"+      - run: npm ci",
	}
	for _, c := range cases {
		if got := SupplyChain(mkFile(".github/workflows/ci.yaml", c)); len(got) != 0 {
			t.Errorf("%q: want none, got %+v", c, got)
		}
	}
	if got := SupplyChain(mkFile("docs/ci.yml", "+      - uses: some-org/x@v1")); len(got) != 0 {
		t.Errorf("non-workflow yaml must be ignored, got %+v", got)
	}
}

func TestSupplyChainLockfile(t *testing.T) {
	honest := mkFile("package-lock.json",
		`-      "resolved": "https://registry.npmjs.org/express/-/express-4.19.1.tgz",`,
		`+      "resolved": "https://registry.npmjs.org/express/-/express-4.19.2.tgz",`,
	)
	if got := SupplyChain(honest); len(got) != 0 {
		t.Fatalf("honest npmjs pin must be silent, got %+v", got)
	}
	bad := mkFile("sub/package-lock.json", `+      "resolved": "https://evil.example.invalid/express.tgz",`)
	got := SupplyChain(bad)
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %+v", got)
	}
	wantFinding(t, got[0], "Lockfile resolved to unexpected registry", finding.RegistryHost, 1, "CWE-829", finding.High, finding.Block)
}

func TestLocalActionIsNotThirdParty(t *testing.T) {
	f := fileOf(".github/workflows/selftest.yml", 1,
		"+      - uses: ./",
		"+      - uses: ./.github/actions/setup",
		"+      - uses: someone/thing@v1",
	)
	got := SupplyChain(f)
	if len(got) != 1 || got[0].Line != 3 {
		t.Fatalf("only the third-party tag must fire, got %+v", got)
	}
}
