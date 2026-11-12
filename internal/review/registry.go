package review

import "github.com/tiagovilasboas/ai-code-review/internal/rules"

// Default is the production rule set. Run dedupes by path:line:cwe with the
// first finding winning, so this order decides ties; the golden reports pin it.
var Default = []rules.Rule{
	rules.SkipProse(rules.XSS),
	rules.SkipProse(rules.SSRF),
	rules.SupplyChain,
	rules.SkipProse(rules.AuthZ),
	rules.Secrets,
	rules.UntrustedDiff,
	rules.SkipProse(rules.SQLInjection),
	rules.SkipProse(rules.CmdInjection),
	rules.SkipProse(rules.PathTraversal),
	rules.Cloudflare,
}
