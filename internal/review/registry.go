package review

import "github.com/tiagovilasboas/ai-code-review/internal/rules"

// Default is the production rule set, in the same order as the JS engine.
var Default = []rules.Rule{
	rules.XSS,
	rules.SSRF,
	rules.SupplyChain,
	rules.AuthZ,
	rules.Secrets,
	rules.UntrustedDiff,
}
