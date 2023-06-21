package githubrelease

import (
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [s][c] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] URL specifies the default github url in case of GitHub enterprise
	URL string `yaml:",omitempty"`
	// [s][c] Username specifies the username used to authenticate with GitHub API
	Username string `yaml:",omitempty"`
	// [s] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [s][c] TypeFilter specifies the GitHub Release type to retrieve before applying the versionfilter rule
	TypeFilter github.ReleaseType `yaml:",omitempty"`
	// [c] Tag allows to check for a specific release tag, default to source output
	Tag string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Owner:      s.Owner,
		Repository: s.Repository,
		URL:        s.URL,
		Tag:        s.Tag,
	}
}
