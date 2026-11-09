package review

import "github.com/tiagovilasboas/ai-code-review/internal/rules"

// Default is the production rule set: the JS engine order, then the Go-only
// injection rules.
var Default = []rules.Rule{
	rules.XSS,
	rules.SSRF,
	rules.SupplyChain,
	rules.AuthZ,
	rules.Secrets,
	rules.UntrustedDiff,
	rules.SQLInjection,
	rules.CmdInjection,
	rules.PathTraversal,
}
