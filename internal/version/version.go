// Package version provides version information for fyaml.
package version

import "fmt"

var (
	// Version is the version of fyaml. It is set at build time via ldflags.
	Version = "dev"
	// Commit is the git commit hash. It is set at build time via ldflags.
	Commit = "unknown"
	// Date is the build date. It is set at build time via ldflags.
	Date = "unknown"
)

// String returns the version string.
func String() string {
	return Version
}

// Full returns the full version string including commit and date.
func Full() string {
	return fmt.Sprintf("%s (commit: %s, date: %s)", Version, Commit, Date)
}
