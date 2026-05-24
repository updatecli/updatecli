package gomodule

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gomodule" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Proxy may have the schemes https, http. file is not supported at this time. If a URL has no scheme, https is assumed
	// [S][C] Proxy allows to override GO proxy similarly to GOPROXY environment variable.
	Proxy string `yaml:",omitempty"`
	// [S][C] Module specifies the name of the module
	Module string `yaml:",omitempty" jsonschema:"required"`
	// [C] Defines a specific package version
	Version string `yaml:",omitempty"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] Age defines the minimum age of a release to be considered valid. It accepts a duration string (e.g., "24h", "7d").
	Age age.Spec `yaml:",omitempty"`
}
