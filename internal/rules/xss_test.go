package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestXSSFixtureCommentBody(t *testing.T) {
	f := mkFile("web/src/components/CommentBody.tsx",
		` import { useComment } from "../hooks/useComment";`,
		` `,
		` type CommentBodyProps = {`,
		`   commentId: string;`,
		` };`,
		` `,
		` export function CommentBody({ commentId }: CommentBodyProps): JSX.Element | null {`,
		`   const comment = useComment(commentId);`,
		` `,
		`   if (!comment) {`,
		`     return null;`,
		`   }`,
		` `,
		`   return (`,
		`     <article className="comment-body" data-comment-id={commentId}>`,
		`+      <div`,
		`+        className="comment-html"`,
		`+        dangerouslySetInnerHTML={{ __html: comment.body }}`,
		`+      />`,
		`     </article>`,
	)
	got := XSS(f)
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %d: %+v", len(got), got)
	}
	wantFinding(t, got[0], "DOM XSS", finding.XSS, 18, "CWE-79", finding.High, finding.Block)
	if got[0].OWASP[0] != "A03:2021" || got[0].OWASP[1] != "ASVS-5.0-1.3.1" {
		t.Fatalf("owasp = %v", got[0].OWASP)
	}
}

func TestXSSSinks(t *testing.T) {
	sinks := []string{
		`el.innerHTML = x`, `el.outerHTML = x`, `el.insertAdjacentHTML("beforeend", x)`,
		`document.write(x)`, `<div v-html="x">`, `this.s.bypassSecurityTrustHtml(x)`, `unsafeHTML(x)`,
	}
	for _, s := range sinks {
		if got := XSS(mkFile("a.js", "+"+s)); len(got) != 1 {
			t.Errorf("%q: want 1 finding, got %d", s, len(got))
		}
	}
}

func TestXSSSkipsCommentsSanitizedAndContext(t *testing.T) {
	cases := [][]string{
		{`+// el.innerHTML = x`},
		{`+<div dangerouslySetInnerHTML={{ __html: sanitize(x) }} />`},
		{`+<div dangerouslySetInnerHTML={{ __html: sanitizeHtml(x) }} />`},
		{` el.innerHTML = x`},
		{`+el.textContent = x`},
	}
	for _, c := range cases {
		if got := XSS(mkFile("a.js", c...)); len(got) != 0 {
			t.Errorf("%q: want no findings, got %+v", c, got)
		}
	}
}
