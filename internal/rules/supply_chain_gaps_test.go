package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func wantRegistryAt(t *testing.T, got []finding.Finding, line int) {
	t.Helper()
	if len(got) != 1 || got[0].Class != finding.RegistryHost || got[0].Line != line {
		t.Fatalf("want one registry-host at line %d, got %+v", line, got)
	}
}

func TestLockfileIntegritySwapSameVersion(t *testing.T) {
	npm := mkFile("package-lock.json",
		` "node_modules/ms": {`, `   "version": "1.3.0",`,
		`   "resolved": "https://registry.npmjs.org/ms/-/ms-1.3.0.tgz",`,
		`+  "integrity": "sha512-BBBB==",`, `   "license": "MIT"`)
	wantRegistryAt(t, SupplyChain(npm), 4)

	yarn := mkFile("yarn.lock",
		` ms@^1.3.0:`, `   version "1.3.0"`, `   resolved "https://registry.yarnpkg.com/ms/-/ms-1.3.0.tgz#abc"`,
		`+  integrity sha512-BBBB==`)
	wantRegistryAt(t, SupplyChain(yarn), 4)

	pnpm := mkFile("pnpm-lock.yaml",
		` `, `   /ms@2.1.3:`, `+    resolution: {integrity: sha512-BBBB==}`, `     dev: false`)
	wantRegistryAt(t, SupplyChain(pnpm), 3)
}

func TestLockfileIntegrityWithVersionBumpIsSilent(t *testing.T) {
	cases := []struct {
		name string
		file string
		in   []string
	}{
		{"npm", "package-lock.json", []string{
			` "node_modules/ms": {`, `+  "version": "1.4.0",`,
			`+  "resolved": "https://registry.npmjs.org/ms/-/ms-1.4.0.tgz",`, `+  "integrity": "sha512-BBBB==",`}},
		{"yarn", "yarn.lock", []string{
			` ms@^1.3.0:`, `+  version "1.4.0"`, `+  resolved "https://registry.yarnpkg.com/ms/-/ms-1.4.0.tgz#abc"`,
			`+  integrity sha512-BBBB==`}},
		{"pnpm", "pnpm-lock.yaml", []string{
			` `, `+  /ms@2.1.4:`, `+    resolution: {integrity: sha512-BBBB==}`}},
		{"new entry", "package-lock.json", []string{
			`+"node_modules/new": {`, `+  "version": "1.0.0",`, `+  "integrity": "sha512-BBBB=="`}},
	}
	for _, c := range cases {
		if got := SupplyChain(mkFile(c.file, c.in...)); len(got) != 0 {
			t.Fatalf("%s: want silent, got %+v", c.name, got)
		}
	}
}

func TestPipIndexHosts(t *testing.T) {
	wantRegistryAt(t, SupplyChain(mkFile("requirements.txt",
		` flask==3.0.0`, `+--extra-index-url https://pkgs.example.invalid/simple/`)), 2)
	wantRegistryAt(t, SupplyChain(mkFile("requirements-dev.txt",
		`+--index-url=https://mirror.example.invalid/simple`)), 1)
	wantRegistryAt(t, SupplyChain(mkFile("requirements.txt",
		`+-i https://mirror.example.invalid/simple`)), 1)
	wantRegistryAt(t, SupplyChain(mkFile("pip.conf",
		` [global]`, `+extra-index-url = https://pkgs.example.invalid/simple/`)), 2)
	wantRegistryAt(t, SupplyChain(mkFile("requirements.txt",
		`+--extra-index-url https://pypi.org@evil.example.invalid/simple`)), 1)

	for _, line := range []string{
		`+--index-url https://pypi.org/simple`,
		`+--extra-index-url https://files.pythonhosted.org/simple`,
		`+# --extra-index-url https://pkgs.example.invalid/simple/`,
		`+--extra-index-url ${PIP_INDEX}`,
		`+flask==3.0.0`,
	} {
		if got := SupplyChain(mkFile("requirements.txt", line)); len(got) != 0 {
			t.Fatalf("%q: want silent, got %+v", line, got)
		}
	}
}

func TestYarnBerryRegistryAndProxy(t *testing.T) {
	wantRegistryAt(t, SupplyChain(mkFile(".yarnrc.yml",
		` nodeLinker: node-modules`, `+npmRegistryServer: "https://pkgs.example.invalid/npm/"`)), 2)
	wantRegistryAt(t, SupplyChain(mkFile(".yarnrc.yml",
		`+npmScopes:`, `+  acme:`, `+    npmRegistryServer: "https://npm.acme.example.invalid/"`)), 3)
	wantRegistryAt(t, SupplyChain(mkFile(".yarnrc.yml",
		`+httpsProxy: "http://proxy.example.invalid:3128"`)), 1)
	wantRegistryAt(t, SupplyChain(mkFile(".yarnrc.yml",
		`+httpProxy: http://proxy.example.invalid:3128`)), 1)

	for _, line := range []string{
		`+npmRegistryServer: "https://registry.yarnpkg.com"`,
		`+    npmRegistryServer: 'https://registry.npmjs.org/'`,
		`+# httpProxy: "http://proxy.example.invalid:3128"`,
		`+nodeLinker: node-modules`,
	} {
		if got := SupplyChain(mkFile(".yarnrc.yml", line)); len(got) != 0 {
			t.Fatalf("%q: want silent, got %+v", line, got)
		}
	}
	if got := SupplyChain(mkFile(".yarnrc.yml", `+npmRegistryServer: "https://registry.npmjs.org.evil.example.invalid"`)); len(got) != 1 {
		t.Fatalf("lookalike host must be flagged, got %+v", got)
	}
}

func TestDockerActionDigestPinning(t *testing.T) {
	digest := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	pinned := mkFile(".github/workflows/ci.yml", `+      - uses: docker://ghcr.io/org/tool@sha256:`+digest)
	if got := SupplyChain(pinned); len(got) != 0 {
		t.Fatalf("digest-pinned docker action must be silent, got %+v", got)
	}
	for _, uses := range []string{
		`docker://ghcr.io/org/tool:latest`,
		`docker://ghcr.io/org/tool@sha256:0123abcd`,
		`docker://ghcr.io/org/tool:1.2@sha256:` + digest[:63],
	} {
		got := SupplyChain(mkFile(".github/workflows/ci.yml", `+      - uses: `+uses))
		if len(got) != 1 || got[0].Class != finding.UnpinnedAction {
			t.Fatalf("%s: want unpinned-action, got %+v", uses, got)
		}
	}
}
