package review

import "github.com/tiagovilasboas/ai-code-review/internal/rules"

// sink wraps a request-reachable code-sink rule: it ignores prose, build
// artifacts and test files. Secrets and untrusted-diff are not wrapped.
func sink(r rules.Rule) rules.Rule {
	return rules.SkipProse(rules.SkipGenerated(rules.SkipTests(r)))
}

// Default is the production rule set. Run dedupes by path:line:cwe with the
// first finding winning, so this order decides ties; the golden reports pin it.
var Default = []rules.Rule{
	sink(rules.XSS),
	sink(rules.SSRF),
	rules.SupplyChain,
	rules.SkipProse(rules.SkipGenerated(rules.AuthZ)),
	rules.Secrets,
	rules.UntrustedDiff,
	sink(rules.SQLInjection),
	sink(rules.CmdInjection),
	sink(rules.PathTraversal),
	rules.Cloudflare,
}
