package review

import "github.com/tiagovilasboas/ai-code-review/internal/rules"

// Default is the production rule set. The root wires XSS, SSRF, SupplyChain,
// AuthZ, Secrets, UntrustedDiff here, in that order.
var Default = []rules.Rule{}
