package tag

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Bitbucket release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] Tag defines the Bitbucket tag .
	Tag string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Owner:      s.Owner,
		Repository: s.Repository,
		Tag:        s.Tag,
	}
}
