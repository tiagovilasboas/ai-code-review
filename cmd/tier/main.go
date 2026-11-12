// Command tier prints the risk tier of a change and how many humans must
// approve it. See docs/review-tier.en.md.
package main

import (
	"os"

	"github.com/tiagovilasboas/ai-code-review/internal/tier"
)

func main() {
	os.Exit(tier.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
