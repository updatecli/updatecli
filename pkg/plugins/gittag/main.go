package gittag

import (
	"github.com/updatecli/updatecli/pkg/plugins/version"
)

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	Path          string         // Path contains the git repository path
	VersionFilter version.Filter // VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	Message       string         // Message associated to the git Tag
}

// GitTag defines a resource of kind "gittag"
type GitTag struct {
	spec         Spec
	foundVersion version.Version // Holds both parsed version and original version (to allow retrieving metadata such as changelog)
}

// New returns a reference to a newly initialized GitTag object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec Spec) (*GitTag, error) {
	newResource := &GitTag{
		spec: spec,
	}

	err := newResource.Validate()
	if err != nil {
		return nil, err
	}

	return newResource, nil
}

// Validate tests that tag struct is correctly configured
func (gt *GitTag) Validate() error {
	err := gt.spec.VersionFilter.Validate()
	if err != nil {
		return err
	}

	return nil
}
