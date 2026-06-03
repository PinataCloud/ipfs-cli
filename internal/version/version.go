// Package version holds build-time version information for the CLI.
package version

// Version is the CLI version, injected at build time via ldflags:
//
//	-X pinata/internal/version.Version=<version>
var Version = "dev"

// UserAgent returns the User-Agent string sent with outgoing HTTP requests.
func UserAgent() string {
	return "pinata-cli/" + Version
}
