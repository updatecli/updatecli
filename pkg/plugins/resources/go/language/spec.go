package language

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"golang" defines the specification for retrieving Go releases.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "version" defines the Go version to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	// example:
	//   * version: 1.23.0
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
	//     pattern: "~1.23"
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
	//   * when set, Updatecli reads the release date of each Go version from the git repository
	//     https://github.com/golang/go.git, using the commit date of each tag.
	//   * fetching the git tags and their commit dates is slow, so use it with care.
	//
	// example:
	// ```
	//   age:
	//     minimum: 7d
	// ```
	//
	Age age.Spec `yaml:",omitempty"`
}
