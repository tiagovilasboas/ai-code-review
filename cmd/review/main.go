// Command review prints the evidence-based AppSec review of a unified diff.
package main

import (
	"os"

	"github.com/tiagovilasboas/ai-code-review/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
