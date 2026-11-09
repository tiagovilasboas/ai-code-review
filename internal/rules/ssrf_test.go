package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestSSRFFixturePreview(t *testing.T) {
	f := mkFile("src/api/preview.ts",
		` import { Request, Response } from "express";`,
		` `,
		` export async function pingStatus(_req: Request, res: Response) {`,
		`   const response = await fetch("https://status.example.invalid/health");`,
		`   return res.json({ ok: response.ok });`,
		` }`,
		` `,
		`+export async function previewLink(req: Request, res: Response) {`,
		`+  const targetUrl = req.body.url;`,
		`+  // Denylist-only smoke target: localhost string match is not an allowlist`,
		`+  if (String(targetUrl).includes("localhost") || String(targetUrl).includes("127.0.0.1")) {`,
		`+    return res.status(400).json({ error: "blocked" });`,
		`+  }`,
		`+  const response = await fetch(targetUrl, { redirect: "follow" });`,
		`+  const html = await response.text();`,
		`+  return res.json({ html });`,
		`+}`,
		`+`,
		` export async function listPreviews(req: Request, res: Response) {`,
	)
	got := SSRF(f)
	if len(got) != 2 {
		t.Fatalf("want 2 findings, got %d: %+v", len(got), got)
	}
	wantFinding(t, got[0], "SSRF / open URL fetch", finding.SSRF, 14, "CWE-918", finding.High, finding.Block)
	wantFinding(t, got[1], "Fetched body returned to client", finding.SSRF, 16, "CWE-918", finding.Medium, finding.Review)
	if got[0].OWASP[0] != "A10:2021" || got[0].OWASP[1] != "ASVS-5.0-1.3.6" {
		t.Fatalf("owasp = %v", got[0].OWASP)
	}
}

func TestSSRFLiteralURLIsNotAFinding(t *testing.T) {
	f := mkFile("src/api/ping.ts", "+const r = await fetch('https://status.example.invalid/health');")
	if got := SSRF(f); len(got) != 0 {
		t.Fatalf("want none, got %+v", got)
	}
}

func TestSSRFClientSideFileIsSkipped(t *testing.T) {
	f := mkFile("web/src/Preview.tsx", "+const r = await fetch(url);")
	if got := SSRF(f); len(got) != 0 {
		t.Fatalf("want none, got %+v", got)
	}
}

func TestSSRFDetectsServerByContent(t *testing.T) {
	f := mkFile("lib/proxy.js",
		` const app = require('express')();`,
		`+const r = await axios.get(target, { timeout: 1 });`,
	)
	got := SSRF(f)
	if len(got) != 1 || got[0].Line != 2 {
		t.Fatalf("want one finding at line 2, got %+v", got)
	}
}

func TestSSRFSkipsComments(t *testing.T) {
	f := mkFile("src/server/x.js", "+// fetch(target)")
	if got := SSRF(f); len(got) != 0 {
		t.Fatalf("want none, got %+v", got)
	}
}
