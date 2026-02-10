package gittag

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
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
	// Tag defines the git tag to check for exact match.
	//
	// compatible:
	//   * condition
	//
	// When specified, the condition will check for an exact tag match
	// instead of using versionFilter pattern matching.
	Tag string `yaml:",omitempty"`
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
	//	  when using the ssh protocol, the user must have the right to clone the repository
	//	  based on its local ssh configuration
	//
	//    it's possible to specify git tags without cloning the repository by using the `lsremote` option,
	//    in that case the URL is required and the tags will be retrieved from the remote repository directly without cloning it.
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
	// LsRemote indicates that the resource should only consider remote tags.
	// When set to true, the tags will be sorted alphabetically to align with the behavior of `git ls-remote --refs --tags`.
	// This means that default version filter "latest" will return the latest tag in alphabetical order,
	// you may need to use a different version filter (e.g. semver) to get the expected tag when using lsRemote.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remarks:
	//   Requires the URL field to be set, as it retrieves tags from the remote repository without cloning it.

	LsRemote *bool `yaml:",omitempty"`
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
	// lsRemote indicates that the resource should only consider remote tags.
	lsRemote bool
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

	// To maintain backward compatibility with existing users,
	// the lsRemote field is defaulted to false if not specified in the manifest.
	lsRemote := false
	if newSpec.LsRemote != nil {
		lsRemote = *newSpec.LsRemote
	}

	newResource := &GitTag{
		spec:             newSpec,
		versionFilter:    newFilter,
		nativeGitHandler: &gitgeneric.GoGit{},
		lsRemote:         lsRemote,
	}

	return newResource, nil
}

// Validate tests that tag struct is correctly configured
func (gt *GitTag) Validate() error {
	validationErrors := []string{}

	if gt.spec.Key != "" && gt.spec.Key != "hash" && gt.spec.Key != "name" {
		validationErrors = append(validationErrors, "The only valid values for Key are 'name', 'hash', or empty.")
	}

	if gt.spec.LsRemote != nil && *gt.spec.LsRemote {
		if gt.spec.Path != "" {
			validationErrors = append(validationErrors, "The parameter `path` cannot be used when `lsRemote` is set to true, as `lsremote` is designed to retrieve tags from the remote repository without cloning it.")
		}
		if gt.spec.URL == "" {
			validationErrors = append(validationErrors, "The parameter `url` is required when `lsRemote` is set to true, as it needs a git repository URL to retrieve tags from the remote repository without cloning it.")
		}
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
		Tag:           gt.spec.Tag,
		Key:           gt.spec.Key,
		URL:           redact.URL(gt.spec.URL),
		SourceBranch:  gt.spec.SourceBranch,
	}
}
