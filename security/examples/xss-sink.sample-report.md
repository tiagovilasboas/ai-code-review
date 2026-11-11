# Sample report — XSS HTML sink (fixture)

Not a production incident. Walkthrough of [`security/skills/xss-html.md`](../skills/xss-html.md) against [`xss-sink.sample.diff`](xss-sink.sample.diff).

## Hunt
- Diff introduces React `dangerouslySetInnerHTML` (hunt table: framework escape hatch).
- `__html` is `comment.body` from `useComment(commentId)` — author-supplied comment text, not a compile-time constant.
- No sanitizer (`DOMPurify` or equivalent) in the hunk. Previous line used `{comment.body}` (React text child = HTML-escaped).

## Findings
- `HIGH` | `web/src/components/CommentBody.tsx:18` | author-supplied `comment.body` assigned to `dangerouslySetInnerHTML` with no sanitizer in the hunk | CWE-79

## Evidence
- Sink: `web/src/components/CommentBody.tsx:18` — `__html: comment.body`.
- Source in the same file: `web/src/components/CommentBody.tsx:8` — `useComment(commentId)`. Treat as stored user content for this smoke. The hook body is not in the diff; that does not make the sink safe.
- What changed: minus-hunk at `web/src/components/CommentBody.tsx:16` was `{comment.body}` (React text child). Plus-hunk `16–19` is the escape hatch.

## Verdict
**FAIL**

## Fix direction
If the product only needs plain text: revert to `{comment.body}` (or `textContent`). If it needs rich HTML: sanitize with DOMPurify (or equivalent) *before* the sink and cite that line in the follow-up review. Do not regex-strip `<script>`.

## Out of scope / not claimed
- No stored-vs-DOM classification beyond “comment body rendered as HTML”.
- No CVSS, no “this is in production”, no exploit PoC.
- CSP / cookie flags not in the diff → not findings.
- `useComment` implementation not shown → not a second finding; already folded into source confidence.
