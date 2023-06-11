package gitbranch

import (
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

func (s Spec) Atomic() Spec {
	return Spec{
		Branch: s.Branch,
		Path:   s.Path,
	}
}
