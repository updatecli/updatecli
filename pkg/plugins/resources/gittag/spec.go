package gittag

import (
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
	// Key of the tag object to retrieve, default is tag "name" filters are always against tag name, this only controls the output; Current options are 'name' and 'hash'.
	Key string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Path: s.Path,
	}
}
