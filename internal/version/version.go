// Package version holds the build version, set with
// -ldflags "-X github.com/tiagovilasboas/ai-code-review/internal/version.Version=1.2.3".
package version

// Version is "dev" unless overridden at link time.
var Version = "dev"
