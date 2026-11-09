package rules

import "testing"

func TestInstallRewritesLockfile(t *testing.T) {
	flagged := []string{
		"npm install", "npm i", "npm install typescript@5", "cd web && npm install",
		"yarn", "yarn install", "bun install", "pnpm install --no-frozen-lockfile",
	}
	silent := []string{
		"npm ci", "npm install -g vercel", "npm i --global pnpm", "npm install --no-package-lock",
		"yarn install --frozen-lockfile", "yarn install --immutable", "yarn test",
		"pnpm install", "echo \"npm install\"", "bun install --frozen-lockfile",
	}
	for _, c := range flagged {
		if !installRewritesLockfile(c) {
			t.Errorf("want flagged: %q", c)
		}
	}
	for _, c := range silent {
		if installRewritesLockfile(c) {
			t.Errorf("want silent: %q", c)
		}
	}
}

func TestWorkflowBlockRunScalar(t *testing.T) {
	f := mkFile(".github/workflows/ci.yml",
		` steps:`, `   - name: Deps`, `     run: |`,
		`+      echo hi`, `+      npm install`, `+      npm test`,
	)
	got := SupplyChain(f)
	if len(got) != 1 || got[0].Line != 5 {
		t.Fatalf("want one finding on line 5, got %+v", got)
	}
}

func TestRegistryHostAllowedIsExact(t *testing.T) {
	ok := []string{"https://registry.npmjs.org/", "https://Registry.NPMJS.org:443/x", "https://registry.yarnpkg.com/a", "${NPM_REGISTRY}"}
	bad := []string{
		"https://registry.npmjs.org.mirror.example.invalid/",
		"https://registry.npmjs.org@evil.example.invalid/",
		"https://evil.example.invalid/registry.npmjs.org/",
		"http://[::1",
	}
	for _, h := range ok {
		if !registryHostAllowed(h) {
			t.Errorf("want allowed: %q", h)
		}
	}
	for _, h := range bad {
		if registryHostAllowed(h) {
			t.Errorf("want rejected: %q", h)
		}
	}
}

func TestSupplyChainLockfileAndDependencySources(t *testing.T) {
	yarn := mkFile("yarn.lock", `+  resolved "https://cdn.example.invalid/a.tgz#abc"`, `+  resolved "https://registry.yarnpkg.com/a/-/a-1.tgz#abc"`)
	if got := SupplyChain(yarn); len(got) != 1 || got[0].Line != 1 {
		t.Fatalf("yarn.lock: %+v", got)
	}
	pnpm := mkFile("pnpm-lock.yaml", `+      tarball: https://cdn.example.invalid/a.tgz`)
	if got := SupplyChain(pnpm); len(got) != 1 {
		t.Fatalf("pnpm-lock: %+v", got)
	}
	pkg := mkFile("package.json",
		`+  "homepage": "https://site.example.invalid",`,
		`+  "url": "git+https://host.example.invalid/o/r.git",`,
		`+  "ui": "https://files.example.invalid/ui.tgz",`,
	)
	if got := SupplyChain(pkg); len(got) != 1 || got[0].Line != 3 {
		t.Fatalf("package.json: %+v", got)
	}
}

func TestSupplyChainInstallScriptScope(t *testing.T) {
	if got := SupplyChain(mkFile(".npmrc", `+ignore-scripts = false`)); len(got) != 1 {
		t.Fatalf("npmrc ignore-scripts=false: %+v", got)
	}
	if got := SupplyChain(mkFile(".npmrc", `+ignore-scripts=true`, `+; ignore-scripts=false`)); len(got) != 0 {
		t.Fatalf("want silent: %+v", got)
	}
	fixture := mkFile("test/fixtures/p/package.json", `+  "scripts": { "postinstall": "echo" }`)
	if got := SupplyChain(fixture); len(got) != 0 {
		t.Fatalf("fixture manifest must be silent: %+v", got)
	}
}
