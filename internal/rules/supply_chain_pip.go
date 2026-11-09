package rules

import (
	"path"
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

var scPipHosts = map[string]bool{
	"pypi.org":               true,
	"files.pythonhosted.org": true,
}

var (
	scPipReqFile   = regexp.MustCompile(`^requirements[\w.-]*\.txt$`)
	scPipConfFile  = regexp.MustCompile(`^pip\.(?:conf|ini)$`)
	scPipReqIndex  = regexp.MustCompile(`^\s*(?:--extra-index-url|--index-url|-i)(?:\s+|=)(\S+)`)
	scPipConfIndex = regexp.MustCompile(`^\s*(?:extra-index-url|index-url)\s*[=:]\s*(\S*)`)
	scPipConfCont  = regexp.MustCompile(`^\s+(\w+://\S+)\s*$`)
)

func isPipFile(p string) bool {
	name := path.Base(p)
	return scPipReqFile.MatchString(name) || scPipConfFile.MatchString(name) ||
		(strings.HasSuffix(name, ".txt") && strings.Contains("/"+p, "/requirements/"))
}

// pipIndexFindings flags pip index options that point outside PyPI, in
// requirements files (--index-url, --extra-index-url, -i) and pip.conf.
func pipIndexFindings(f diff.File) []finding.Finding {
	conf := scPipConfFile.MatchString(path.Base(f.Path))
	var out []finding.Finding
	for _, l := range f.Added {
		if isNpmrcComment(l.Text) {
			continue
		}
		re := scPipReqIndex
		if conf {
			re = scPipConfIndex
		}
		m := re.FindStringSubmatch(l.Text)
		if m == nil && conf {
			m = scPipConfCont.FindStringSubmatch(l.Text)
		}
		if m == nil || m[1] == "" || hostAllowedIn(scPipHosts, m[1]) {
			continue
		}
		out = add(out, scInput(f, l.Line, finding.RegistryHost, "Unexpected pip index host", "CWE-829", finding.High))
	}
	return out
}
