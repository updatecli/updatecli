package cargopackage

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// !deprecated, please use Registry.URL
	IndexUrl string `yaml:",omitempty" jsonschema:"-"`
	// [S][C] Registry specifies the registry to use
	Registry cargo.Registry `yaml:",omitempty"`
	// [S][C] Package specifies the name of the package
	Package string `yaml:",omitempty" jsonschema:"required"`
	// [C] Defines a specific package version
	Version string `yaml:",omitempty"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}
