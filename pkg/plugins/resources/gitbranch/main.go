package gitbranch

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"gitbranch" defines the specification for manipulating Git branches.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "path" defines the path of the local Git repository.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * a relative path is resolved from the Updatecli manifest directory.
	//   * "path" overrides the working directory provided by the scm configuration.
	//   * "url" takes precedence over "path".
	//
	Path string `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its type, such as regex, semver, or latest.
	//
	// compatible:
	//   * source
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a branch to be considered valid.
	//
	// compatible:
	//   * source
	//
	// remark:
	//   * "minimum" and "maximum" accept a duration string such as "24h", "7d", "3w" or "1y".
	//   * the age of a branch is the date of its latest commit.
	//   * when the age filter discards every branch, the source is skipped.
	//
	// example:
	// ```
	//   age:
	//     minimum: 7d
	// ```
	//
	Age age.Spec `yaml:",omitempty"`
	// "branch" defines the Git branch name.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	Branch string `yaml:",omitempty"`
	// "depth" limits the number of commits fetched from the Git repository.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   0, which means no limit
	//
	// remark:
	//   * Updatecli cannot find branches that are not included in the fetched commits.
	//
	Depth *int `yaml:",omitempty"`
	// "sourcebranch" defines the branch used as the starting point of the new Git branch.
	//
	// compatible:
	//   * target
	//
	// remark:
	//   * "sourcebranch" is required when no "scmid" is set.
	//
	SourceBranch string `yaml:",omitempty"`
	// "url" defines the Git repository URL to clone.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * with the ssh protocol, the user must be allowed to clone the repository
	//     with their local ssh configuration.
	//   * "url" overrides "path" and the working directory provided by the scm configuration.
	//
	// example:
	//   * url: git@github.com:updatecli/updatecli.git
	//   * url: https://github.com/updatecli/updatecli.git
	//
	URL string `yaml:",omitempty" jsonschema:"required"`
	// "username" defines the username used with the HTTP protocol.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	Username string `yaml:",omitempty"`
	// "password" defines the password used with the HTTP protocol.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	Password string `yaml:",omitempty"`
	// "key" defines which attribute of the branch the source returns.
	//
	// compatible:
	//   * source
	//
	// default:
	//   name
	//
	// remark:
	//   * accepted values are "name", "hash" or empty.
	//
	// example:
	//   * key: hash
	//
	Key string `yaml:",omitempty"`
}

// GitBranch defines a resource of kind "gitbranch"
type GitBranch struct {
	spec Spec
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// nativeGitHandler holds a git client implementation to manipulate git SCMs
	nativeGitHandler gitgeneric.GitHandler
	// branch hold the branch used for condition and target
	branch string
	// directory defines the local path where the git repository is cloned.
	directory string
}

// New returns a reference to a newly initialized GitBranch object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec interface{}) (*GitBranch, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	validationErrors := []string{}

	if newSpec.Key != "" && newSpec.Key != "hash" && newSpec.Key != "name" {
		validationErrors = append(validationErrors, "The only valid values for Key are 'name', 'hash', or empty.")
	}

	if err := newSpec.Age.Validate(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return &GitBranch{}, fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitBranch{}, err
	}

	newResource := &GitBranch{
		spec:             newSpec,
		versionFilter:    newFilter,
		nativeGitHandler: &gitgeneric.GoGit{},
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (gb *GitBranch) Changelog(from, to string) *result.Changelogs {
	return nil
}

// clone clones the git repository
func (gb *GitBranch) clone() (string, error) {
	g, err := git.New(git.Spec{
		URL:      gb.spec.URL,
		Username: gb.spec.Username,
		Password: gb.spec.Password,
		Depth:    gb.spec.Depth,
	}, "")
	if err != nil {
		return "", err
	}
	return g.Clone()
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information
func (gb *GitBranch) ReportConfig() interface{} {
	return Spec{
		Path:          gb.spec.Path,
		Branch:        gb.spec.Branch,
		VersionFilter: gb.spec.VersionFilter,
		Age:           gb.spec.Age,
		SourceBranch:  gb.spec.SourceBranch,
		// Ensure that the URL doesn't contain any sensitive information
		URL: redact.URL(gb.spec.URL),
		Key: gb.spec.Key,
	}
}
