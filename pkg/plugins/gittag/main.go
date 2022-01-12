package gittag

import (
	"github.com/sirupsen/logrus"
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

// Validate tests that tag struct is correctly configured
func (gt *GitTag) Validate() error {
	err := gt.spec.VersionFilter.Validate()
	if err != nil {
		return err
	}

	if len(gt.spec.Message) == 0 {
		logrus.Warningf("no git tag message specified")
		gt.spec.Message = "Generated by updatecli"
	}
	return nil
}
