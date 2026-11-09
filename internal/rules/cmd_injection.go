package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// cmdCtx is what a command sink check needs to know about its file.
type cmdCtx struct {
	php     bool
	call    string          // the sink call, possibly multi-line
	built   map[string]bool // variables holding a string built from non-literals
	tainted map[string]bool
	safeCmd map[string]bool // PHP variables holding only escapeshellarg results
}

// cmdSink is one shell-reaching call and the check that makes it a finding.
type cmdSink struct {
	lang  lang
	re    *regexp.Regexp
	check func(args []string, c cmdCtx) bool
}

var (
	childProcess = regexp.MustCompile(`\bchild_process\b`)
	jsShellTrue  = regexp.MustCompile(`\bshell\s*:\s*true\b`)
	pyShellTrue  = regexp.MustCompile(`\bshell\s*=\s*True\b`)
	phpBacktick  = regexp.MustCompile("`[^`]*\\$[A-Za-z_][^`]*`")
	shellNames   = map[string]bool{"sh": true, "bash": true, "zsh": true, "/bin/sh": true, "/bin/bash": true,
		"cmd": true, "cmd.exe": true, "powershell": true, "pwsh": true}
	shellFlags = map[string]bool{"-c": true, "/c": true, "/C": true, "-Command": true}
)

var cmdSinks = []cmdSink{
	{langJS, regexp.MustCompile(`(?:^|[^.\w$])(?:exec|execSync)\s*\(|\b(?:child_process|childProcess|cp)\.(?:exec|execSync)\s*\(`), builtOrTainted},
	{langJS, regexp.MustCompile(`\b(?:spawn|spawnSync|execFile|execFileSync)\s*\(`), jsShellSpawn},
	{langPython, regexp.MustCompile(`\bsubprocess\.(?:run|call|Popen|check_output|check_call)\s*\(`), pyShellSubprocess},
	{langPython, regexp.MustCompile(`\bos\.(?:system|popen)\s*\(|\bsubprocess\.(?:getoutput|getstatusoutput)\s*\(`), firstNonLiteral},
	{langGo, regexp.MustCompile(`\bexec\.Command(?:Context)?\s*\(`), shellDashC},
	{langJava, regexp.MustCompile(`\.getRuntime\(\)\s*\.exec\s*\(`), javaRuntimeExec},
	{langJava, regexp.MustCompile(`\bnew\s+ProcessBuilder\s*\(`), shellDashC},
	{langPHP, regexp.MustCompile(`(?:^|[^\w$>:])(?:shell_exec|system|passthru|exec|popen|proc_open|pcntl_exec)\s*\(`), phpShellCall},
}

// CmdInjection flags shell commands built from non-literal values (CWE-78).
// It is HIGH when request input visibly reaches the command, MEDIUM otherwise.
// Literal commands and argv-style calls without a shell never fire.
func CmdInjection(f diff.File) []finding.Finding {
	lg := langOf(f.Path)
	if lg == langOther {
		return nil
	}
	c := cmdCtx{php: lg == langPHP}
	c.tainted = taintedNames(f.Lines, c.php)
	c.built = builtStringVars(f.Lines, c.php, c.tainted)
	if c.php {
		c.safeCmd = phpEscapedVars(f.Lines)
	}
	nodeShell := lg == langJS && anyLineMatches(f.Lines, childProcess)
	var out []finding.Finding
	for i, l := range f.Lines {
		if l.Kind != diff.Added || IsCommentLine(l.Text) {
			continue
		}
		hit, fromRequest := false, false
		for _, s := range cmdSinks {
			loc := s.re.FindStringIndex(l.Text)
			if s.lang != lg || loc == nil || (lg == langJS && !nodeShell) {
				continue
			}
			c.call = callFrom(f.Lines, i, loc[0])
			args := callArgs(c.call, loc[1]-1-loc[0])
			if s.check(args, c) {
				hit, fromRequest = true, c.fromRequest(args)
				break
			}
		}
		if cmd := phpBacktick.FindString(withoutQuoted(l.Text)); !hit && c.php && cmd != "" {
			// PHP backticks run a shell and interpolate like double quotes.
			hit, fromRequest = true, reachesRequest(`"`+cmd[1:len(cmd)-1]+`"`, c.tainted, true)
		}
		if !hit {
			continue
		}
		sev := finding.Medium
		if fromRequest {
			sev = finding.High
		}
		out = add(out, finding.Input{
			Title:    "Shell command built from non-literal input",
			Class:    finding.CmdInjection,
			Path:     f.Path,
			Line:     l.Line,
			CWE:      "CWE-78",
			OWASP:    finding.LabelsFor(finding.CmdInjection),
			Severity: sev,
		})
	}
	return out
}

// fromRequest reports whether request input reaches the command arguments.
func (c cmdCtx) fromRequest(args []string) bool {
	joined := strings.Join(args, ",")
	if reachesRequest(joined, c.tainted, c.php) {
		return true
	}
	code := scanExpr(joined, c.php).code
	for name, tainted := range c.built {
		if tainted && mentions(code, name) {
			return true
		}
	}
	return false
}

// builtOrTainted: the command string is built from non-literals or is request input.
func builtOrTainted(args []string, c cmdCtx) bool {
	if len(args) == 0 {
		return false
	}
	e := scanExpr(args[0], c.php)
	return len(e.dynamicOperands(c.php)) > 0 || mentionsAny(e.code, c.built) || reachesRequest(args[0], c.tainted, c.php)
}

// jsShellSpawn: spawn/execFile only reach a shell with shell: true.
func jsShellSpawn(args []string, c cmdCtx) bool {
	if !jsShellTrue.MatchString(c.call) {
		return false
	}
	for _, a := range args {
		if builtOrTainted([]string{a}, c) {
			return true
		}
	}
	return false
}

// pyShellSubprocess: subprocess with shell=True and a non-literal command.
func pyShellSubprocess(args []string, c cmdCtx) bool {
	return pyShellTrue.MatchString(c.call) && firstNonLiteral(args, c)
}

// firstNonLiteral: the call always runs a shell, so any non-literal command counts.
func firstNonLiteral(args []string, c cmdCtx) bool {
	return len(args) > 0 && len(scanExpr(args[0], c.php).nonLiteralOperands()) > 0
}

// shellDashC: ("sh", "-c", X) with a non-literal X, in any argument position.
func shellDashC(args []string, c cmdCtx) bool {
	for k := 0; k+1 < len(args); k++ {
		if !shellNames[unquote(args[k])] || !shellFlags[unquote(args[k+1])] {
			continue
		}
		for _, a := range args[k+2:] {
			if len(scanExpr(a, c.php).nonLiteralOperands()) > 0 {
				return true
			}
		}
	}
	return false
}

// javaRuntimeExec: a concatenated command, or a {"sh", "-c", X} array.
func javaRuntimeExec(args []string, c cmdCtx) bool {
	if builtOrTainted(args, c) {
		return true
	}
	if len(args) == 0 {
		return false
	}
	if open := strings.IndexByte(args[0], '{'); open >= 0 {
		return shellDashC(callArgs(args[0], open), c)
	}
	return false
}

// builtStringVars maps each variable assigned a string built from
// non-literal values to whether request input reaches it.
func builtStringVars(lines []diff.Line, php bool, tainted map[string]bool) map[string]bool {
	built := map[string]bool{}
	for _, l := range lines {
		if IsCommentLine(l.Text) {
			continue
		}
		a, ok := parseAssignment(l.Text)
		if !ok || len(a.names) != 1 {
			continue
		}
		e := scanExpr(a.rhs, php)
		if len(e.dynamicOperands(php)) == 0 {
			if !a.appends() {
				delete(built, a.names[0])
			}
			continue
		}
		built[a.names[0]] = built[a.names[0]] || reachesRequest(a.rhs, tainted, php)
	}
	return built
}
