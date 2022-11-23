package gitbranch

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gitbranch" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [s][c][t] Path contains the git repository path
	Path string `yaml:",omitempty"`
	// [s] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [c][t] Specify branch name
	Branch string `yaml:",omitempty"`
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
}

// New returns a reference to a newly initialized GitBranch object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec interface{}) (*GitBranch, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitBranch{}, err
	}

	newResource := &GitBranch{
		spec:             newSpec,
		versionFilter:    newFilter,
		nativeGitHandler: gitgeneric.GoGit{},
	}

	return newResource, nil
}

// Validate tests that tag struct is correctly configured
func (gt *GitBranch) Validate() error {
	validationErrors := []string{}
	if gt.spec.Path == "" {
		validationErrors = append(validationErrors, "Git working directory path is empty while it must be specified. Did you specify an `scmID` or a `spec.path`?")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (gt *GitBranch) Changelog() string {
	return ""
}
