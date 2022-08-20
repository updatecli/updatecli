package gittag

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Path contains the git repository path
	Path string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// Message associated to the git tag
	Message string `yaml:",omitempty"`
	// Keeporiginalversion is an ephemeral parameters. cfr https://github.com/updatecli/updatecli/issues/803
	KeepOriginalVersion bool `yaml:",omitempty"`
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
		nativeGitHandler: gitgeneric.GoGit{},
	}

	return newResource, nil
}

// Validate tests that tag struct is correctly configured
func (gt *GitTag) Validate() error {
	validationErrors := []string{}
	if gt.spec.Path == "" {
		validationErrors = append(validationErrors, "Git working directory path is empty while it must be specified. Did you specify an `scmID` or a `spec.path`?")
	}

	if !gt.spec.KeepOriginalVersion && gt.spec.VersionFilter.Kind == version.SEMVERVERSIONKIND {
		logrus.Warningf("%s\n\n", DeprecatedSemverVersionMessage)
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (gt *GitTag) Changelog() string {
	return ""
}
