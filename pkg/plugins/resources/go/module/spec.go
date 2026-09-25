package gomodule

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"golang/module" defines the specification for retrieving Go module versions from a Go proxy.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "proxy" defines the Go proxy to query, similar to the GOPROXY environment variable.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   the GOPROXY environment variable when set, otherwise https://proxy.golang.org
	//
	// remark:
	//   * the schemes "https://" and "http://" are supported. "file://" is not supported yet.
	//   * a URL without a scheme uses https.
	//   * several proxies can be listed, separated by commas.
	//
	// example:
	//   * proxy: https://proxy.golang.org
	//
	Proxy string `yaml:",omitempty"`
	// "module" defines the name of the Go module.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * module: github.com/sirupsen/logrus
	//
	Module string `yaml:",omitempty" jsonschema:"required"`
	// "version" defines the module version to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	// example:
	//   * version: v1.9.3
	//
	Version string `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its type, such as regex, semver or latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: semver
	//   pattern: "*"
	//
	// example:
	// ```
	//   versionfilter:
	//     kind: semver
	//     pattern: "~1.9"
	// ```
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a release to be considered valid.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * in a source, when every published version is discarded by the age filter, the source is skipped.
	//
	// example:
	// ```
	//   age:
	//     minimum: 7d
	// ```
	//
	Age age.Spec `yaml:",omitempty"`
}
