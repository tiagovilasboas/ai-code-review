// Package finding defines the fail-closed review finding: no path:line,
// official CWE and official OWASP IDs, no finding.
package finding

import "regexp"

// Severity follows GitHub security severity labels.
type Severity string

const (
	Critical Severity = "CRITICAL"
	High     Severity = "HIGH"
	Medium   Severity = "MEDIUM"
	Low      Severity = "LOW"
)

// Rank orders severities, most severe first.
func (s Severity) Rank() int {
	switch s {
	case Critical:
		return 0
	case High:
		return 1
	case Medium:
		return 2
	default:
		return 3
	}
}

// Action is what the reviewer must do with a finding.
type Action string

const (
	Block  Action = "BLOCK"
	Review Action = "REVIEW"
)

// Class is the finding family. Values match the JS engine and the skill docs.
type Class string

const (
	AuthZ           Class = "authz"
	Secret          Class = "secret"
	CORS            Class = "cors"
	XSS             Class = "xss"
	SSRF            Class = "ssrf"
	InstallScript   Class = "install-script"
	RegistryHost    Class = "registry-host"
	UnpinnedAction  Class = "unpinned-action"
	LockfileRewrite Class = "lockfile-rewrite"
	UntrustedDiff   Class = "untrusted-diff"
	SQLInjection    Class = "sql-injection"
	CmdInjection    Class = "command-injection"
	PathTraversal   Class = "path-traversal"
)

var classLabel = map[Class]string{
	AuthZ:           "AuthZ/IDOR",
	Secret:          "Secret in source",
	CORS:            "CORS misconfiguration",
	XSS:             "XSS / HTML sink",
	SSRF:            "SSRF / egress",
	InstallScript:   "Supply chain",
	RegistryHost:    "Supply chain",
	UnpinnedAction:  "Supply chain",
	LockfileRewrite: "Supply chain",
	UntrustedDiff:   "Untrusted diff instruction",
	SQLInjection:    "SQL injection",
	CmdInjection:    "OS command injection",
	PathTraversal:   "Path traversal",
}

// Label is the human class name printed in reports.
func (c Class) Label() string { return classLabel[c] }

// ChangesMerge reports whether a BLOCK of this class changes the merge decision.
func (c Class) ChangesMerge() bool { return c == AuthZ || c == Secret }

// Finding is one evidence-backed review result.
type Finding struct {
	Title    string   `json:"title"`
	Class    Class    `json:"findingClass"`
	Path     string   `json:"path"`
	Line     int      `json:"line"`
	CWE      string   `json:"cwe"`
	OWASP    []string `json:"owasp"`
	Severity Severity `json:"severity"`
	Action   Action   `json:"action"`
}

// Input is what a rule proposes; New decides whether it becomes a Finding.
type Input struct {
	Title    string
	Class    Class
	Path     string
	Line     int
	CWE      string
	OWASP    []string
	Severity Severity
}

var cweID = regexp.MustCompile(`^CWE-\d+$`)

// New is fail-closed: it returns false for anything without a real new-file
// path:line, an official CWE, a known class and official OWASP IDs.
// CRITICAL and HIGH become BLOCK; everything else is REVIEW.
func New(in Input) (Finding, bool) {
	if in.Path == "" || in.Line < 1 || !cweID.MatchString(in.CWE) {
		return Finding{}, false
	}
	if _, known := classLabel[in.Class]; !known {
		return Finding{}, false
	}
	if len(in.OWASP) == 0 || !allOfficial(in.OWASP) {
		return Finding{}, false
	}
	action := Review
	if in.Severity == Critical || in.Severity == High {
		action = Block
	}
	return Finding{
		Title:    in.Title,
		Class:    in.Class,
		Path:     in.Path,
		Line:     in.Line,
		CWE:      in.CWE,
		OWASP:    append([]string(nil), in.OWASP...),
		Severity: in.Severity,
		Action:   action,
	}, true
}
