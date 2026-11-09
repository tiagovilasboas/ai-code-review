package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

func TestPHPEscapeShellArgBuiltCommand(t *testing.T) {
	runRuleCases(t, CmdInjection, finding.CmdInjection, "CWE-78", []ruleCase{
		{"escapeshellarg in built var", mkFile("public/ping.php",
			`+$cmd = "ping -c 1 " . escapeshellarg($_GET['host']);`,
			`+$out = shell_exec($cmd);`,
		), nil, finding.Medium},
		{"escapeshellarg nested call", mkFile("public/ping.php",
			`+system("ls " . escapeshellarg(trim($_GET['dir'])));`,
		), nil, finding.Medium},
		{"escaped arg var then exec", mkFile("public/ping.php",
			`+$arg = escapeshellarg($_POST['file']);`,
			`+exec("cat " . $arg, $lines);`,
		), nil, finding.Medium},
		{"escapeshellcmd alone still flagged", mkFile("public/ping.php",
			`+$out = shell_exec("convert " . escapeshellcmd($_GET['file']) . " out.png");`,
		), []int{1}, finding.High},
		{"bare var next to escaped one", mkFile("public/ping.php",
			`+$out = shell_exec("ping -c 1 " . escapeshellarg($host) . " " . $_GET['opts']);`,
		), []int{1}, finding.High},
		{"built var with one bare part", mkFile("public/ping.php",
			`+$cmd = "ping " . escapeshellarg($host) . " " . $_GET['opts'];`,
			`+$out = shell_exec($cmd);`,
		), []int{2}, finding.High},
	})
}

func TestWorkflowScriptInjection(t *testing.T) {
	const wf = ".github/workflows/triage.yml"
	runRuleCases(t, CmdInjection, finding.CmdInjection, "CWE-78", []ruleCase{
		{"inline issue title", mkFile(wf,
			`+      - run: echo "${{ github.event.issue.title }}"`,
		), []int{1}, finding.High},
		{"block continuation line", mkFile(wf,
			`+      - name: greet`,
			`+        run: |`,
			`+          set -e`,
			`+          echo "PR: ${{ github.event.pull_request.title }}"`,
			`+          echo done`,
		), []int{4}, finding.High},
		{"head ref", mkFile(wf,
			`+        run: git checkout ${{ github.head_ref }}`,
		), []int{1}, finding.High},
		{"pr head ref", mkFile(wf,
			`+        run: git checkout ${{github.event.pull_request.head.ref}}`,
		), []int{1}, finding.High},
		{"comment body", mkFile(wf,
			`+        run: echo '${{ github.event.comment.body }}' > c.txt`,
		), []int{1}, finding.High},
		{"commits message", mkFile(wf,
			`+        run: echo "${{ github.event.commits[0].message }}"`,
		), []int{1}, finding.High},
		{"review body yaml extension", mkFile(".github/workflows/r.yaml",
			`+        run: echo ${{ github.event.review.body }}`,
		), []int{1}, finding.High},
		{"env indirection", mkFile(wf,
			`+        env:`,
			`+          TITLE: ${{ github.event.issue.title }}`,
			`+        run: echo "$TITLE"`,
		), nil, finding.Medium},
		{"safe expression", mkFile(wf,
			`+        run: echo "${{ github.sha }} ${{ github.event.issue.number }}"`,
		), nil, finding.Medium},
		{"expression in with not run", mkFile(wf,
			`+        with:`,
			`+          title: ${{ github.event.issue.title }}`,
		), nil, finding.Medium},
		{"not a workflow file", mkFile("docs/example.yml",
			`+        run: echo "${{ github.event.issue.title }}"`,
		), nil, finding.Medium},
		{"commented run", mkFile(wf,
			`+        # run: echo "${{ github.event.issue.title }}"`,
		), nil, finding.Medium},
	})
}
