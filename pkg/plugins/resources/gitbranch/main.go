package gitbranch

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gitbranch" resource
// parsed from an updatecli manifest file
type Spec struct {
	// path contains the git repository path
	Path string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	//  compatible:
	//    * source
	//    * condition
	//    * target
	VersionFilter version.Filter `yaml:",omitempty"`
	// branch specifies the branch name
	//
	//  compatible:
	//    * source
	//    * condition
	//    * target
	Branch string `yaml:",omitempty"`
	//	"url" specifies the git url to use for fetching Git Tags.
	//
	//	compatible:
	//	  * source
	//	  * condition
	// 	  * target
	//
	//	example:
	//	  * git@github.com:updatecli/updatecli.git
	//	  * https://github.com/updatecli/updatecli.git
	//
	//	remarks:
	//		when using the ssh protocol, the user must have the right to clone the repository
	//		based on its local ssh configuration
	SourceBranch string `yaml:",omitempty"`
	// "sourcebranch" defines the branch name used as a source to create the new Git branch.
	//
	// compatible:
	//  * target
	//
	// remark:
	//  * sourcebranch is required when the scmid is not defined.
	URL string `yaml:",omitempty" jsonschema:"required"`
	//	"username" specifies the username when using the HTTP protocol
	//
	//	compatible
	//	  * source
	//	  * condition
	// 	  * target
	Username string `yaml:",omitempty"`
	//	"password" specifies the password when using the HTTP protocol
	//
	//	compatible:
	//	  * source
	// 	  * condition
	// 	  * target
	Password string `yaml:",omitempty"`
	//  "key" of the tag object to retrieve.
	//
	//  Accepted values: ['name','hash'].
	//
	//  Default: 'name'
	//  Compatible:
	//    * source
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
		SourceBranch:  gb.spec.SourceBranch,
		// Ensure that the URL doesn't contain any sensitive information
		URL: redact.URL(gb.spec.URL),
		Key: gb.spec.Key,
	}
}
