package gittaghash

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gittaghash" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Path contains the git repository path
	Path string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// Message associated to the git tag
	Message string `yaml:",omitempty"`
}

// GitTagHash defines a resource of kind "gittaghash"
type GitTagHash struct {
	spec Spec
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// nativeGitHandler holds a git client implementation to manipulate git SCMs
	nativeGitHandler gitgeneric.GitHandler
}

// New returns a reference to a newly initialized GitTagHash object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec interface{}) (*GitTagHash, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitTagHash{}, err
	}

	newResource := &GitTagHash{
		spec:             newSpec,
		versionFilter:    newFilter,
		nativeGitHandler: gitgeneric.GoGit{},
	}

	return newResource, nil
}

// Validate tests that tag struct is correctly configured
func (gt *GitTagHash) Validate() error {
	validationErrors := []string{}
	// Catch missing scmid or empty path
	if gt.spec.Path == "" {
		validationErrors = append(validationErrors, "Git working directory path is empty while it must be specified. Did you specify an `scmID` or a `spec.path`?")
	}
	// Gittaghash is specifically for repos which do not follow valid semver naming schemas in their tags
	//  the hash is used to tell the 'go get' tool what to get in a reliable way
	if gt.spec.VersionFilter.Kind == "semver" {
		validationErrors = append(validationErrors, "Semver searching is not supported, use gittag plugin instead.")
	}

	// Return all the validation errors if any found
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (gt *GitTagHash) Changelog() string {
	return ""
}
