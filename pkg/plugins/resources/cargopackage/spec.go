package cargopackage

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C] IndexUrl specifies the url of the index to use to check version
	// If set, package version will be checked against the API
	IndexUrl string `yaml:",omitempty"`
	// [S][C] IndexDir specifies the directory of the index to use to check version
	// If set, package version will be checked using the file structure
	IndexDir string `yaml:",omitempty"`
	// [S][C] Package specifies the name of the package
	Package string `yaml:",omitempty" jsonschema:"required"`
	// [C] Defines a specific package version
	Version string `yaml:",omitempty"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}
