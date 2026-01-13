// Package version provides version information for razd.
package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags.
var (
	// Version is the semantic version of razd
	Version = "dev"
	// Commit is the git commit hash
	Commit = "unknown"
	// Date is the build date
	Date = "unknown"
)

// Info returns the full version string.
func Info() string {
	return fmt.Sprintf("razd version %s (%s) %s/%s",
		Version,
		Commit[:min(7, len(Commit))],
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// Short returns just the version number.
func Short() string {
	return Version
}
