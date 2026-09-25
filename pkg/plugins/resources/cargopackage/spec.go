package cargopackage

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"cargopackage" defines the specification for retrieving a Cargo package version from a registry.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "indexurl" defines the registry URL.
	//
	// deprecated:
	//   * use "registry.url" instead.
	//   * it is ignored when "registry.url" is set.
	//
	IndexUrl string `yaml:",omitempty" jsonschema:"-"`
	// "registry" defines the Cargo registry to query.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   the crates.io API, https://crates.io/api/v1/crates
	//
	// remark:
	//   * in a condition with an scm, the scm directory is used as the registry root directory
	//     and "registry.url" and "registry.rootdir" are ignored.
	//
	Registry cargo.Registry `yaml:",omitempty"`
	// "package" defines the name of the Cargo package.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * it is required.
	//
	// example:
	//   * package: serde
	//
	Package string `yaml:",omitempty" jsonschema:"required"`
	// "version" defines the package version to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	// example:
	//   * version: 1.0.0
	//
	Version string `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its kind, such as regex, semver or latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: latest
	//
	// remark:
	//   * yanked versions are ignored.
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}
