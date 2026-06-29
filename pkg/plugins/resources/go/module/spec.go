package gomodule

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gomodule" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Proxy allows to override GO proxy similarly to GOPROXY environment variable.
	// Proxy may have the schemes https, http. file is not supported at this time. If a URL has no scheme, https is assumed
	// Compatible:
	//   * source
	//   * condition
	//
	Proxy string `yaml:",omitempty"`
	// module specifies the name of the Golang module
	//
	// Compatible:
	//   * source
	//   * condition
	//
	Module string `yaml:",omitempty" jsonschema:"required"`
	// version defines a specific package version to check
	//
	// Compatible:
	//   * condition
	//
	Version string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	// Compatible:
	//   * source
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// Age defines the minimum or maximum age of a release to be considered valid. It accepts a duration string (e.g., "24h", "7d").
	//
	// Compatible:
	//   * source
	//   * condition
	//
	Age age.Spec `yaml:",omitempty"`
}
