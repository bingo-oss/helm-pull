package version

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
)

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion string
)

// Map provides the iterable version information.
var Map = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildUser": BuildUser,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}

const (
	VersionPrefix = ""

	// ExampleVersion shows an example version in the help
	// if no version could be found (which should never really happen!)
	ExampleVersion = "0.0.1"

	// TestVersion used in test cases for the current version if no
	// version can be found - such as if the version property is not properly
	// included in the go test flags
	TestVersion = "0.0.1"
)

func GetVersion() string {
	v := Map["version"]
	if v == "" {
		v = TestVersion
	}
	return v
}

func GetSemverVersion() (semver.Version, error) {
	return semver.Make(strings.TrimPrefix(GetVersion(), VersionPrefix))
}

// VersionStringDefault returns the current version string or returns a dummy
// default value if there is an error
func VersionStringDefault(defaultValue string) string {
	v, err := GetSemverVersion()
	if err == nil {
		return v.String()
	}
	fmt.Printf("Warning failed to load version: %s\n", err)
	return defaultValue
}
