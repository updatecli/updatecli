package gittag

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

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Path contains the git repository path
	Path string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	//  compatible:
	//    * source
	//    * condition
	//    * target
	VersionFilter version.Filter `yaml:",omitempty"`
	//  Message associated to the git tag
	//
	//  compatible:
	//    * target
	Message string `yaml:",omitempty"`
	//  "key" of the tag object to retrieve.
	//
	//  Accepted values: ['name','hash'].
	//
	//  Default: 'name'
	//  Compatible:
	//    * source
	Key string `yaml:",omitempty"`
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
	// "sourcebranch" defines the branch name used as a source to create the new Git branch.
	//
	// compatible:
	//  * target
	//
	// remark:
	//  * sourcebranch is required when the scmid is not defined.
	SourceBranch string `yaml:",omitempty"`
}

// GitTag defines a resource of kind "gittag"
type GitTag struct {
	spec Spec
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// nativeGitHandler holds a git client implementation to manipulate git SCMs
	nativeGitHandler gitgeneric.GitHandler
	// directory defines the local path where the git repository is cloned.
	directory string
}

// New returns a reference to a newly initialized GitTag object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec interface{}) (*GitTag, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitTag{}, err
	}

	newResource := &GitTag{
		spec:             newSpec,
		versionFilter:    newFilter,
		nativeGitHandler: &gitgeneric.GoGit{},
	}

	return newResource, nil
}

// Validate tests that tag struct is correctly configured
func (gt *GitTag) Validate() error {
	validationErrors := []string{}

	if gt.spec.Key != "" && gt.spec.Key != "hash" && gt.spec.Key != "name" {
		validationErrors = append(validationErrors, "The only valid values for Key are 'name', 'hash', or empty.")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (gt *GitTag) Changelog(from, to string) *result.Changelogs {
	return nil
}

// clone clones the git repository
func (gt *GitTag) clone() (string, error) {
	g, err := git.New(git.Spec{
		URL:      gt.spec.URL,
		Username: gt.spec.Username,
		Password: gt.spec.Password,
	}, "")

	if err != nil {
		return "", err
	}
	return g.Clone()
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information
// and context specific data.
func (gt *GitTag) ReportConfig() interface{} {
	return Spec{
		Path:          gt.spec.Path,
		VersionFilter: gt.spec.VersionFilter,
		Key:           gt.spec.Key,
		URL:           redact.URL(gt.spec.URL),
		SourceBranch:  gt.spec.SourceBranch,
	}
}
