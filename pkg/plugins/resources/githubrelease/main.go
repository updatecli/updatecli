package githubrelease

import (
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty" jsonschema:"required"`
	// URL specifies the default github url in case of GitHub enterprise
	URL string `yaml:",omitempty"`
	// Username specifies the username used to authenticate with Github API
	Username string `yaml:",omitempty" jsonschema:"required"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// KeepOriginalVersion is an ephemeral parameters. cfr https://github.com/updatecli/updatecli/issues/803
	KeepOriginalVersion bool
}

// GitHubRelease defines a resource of kind "githubrelease"
type GitHubRelease struct {
	ghHandler           github.GithubHandler
	versionFilter       version.Filter // Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	foundVersion        version.Version
	keepOriginalVersion bool
}

// New returns a new valid GitHubRelease object.
func New(spec interface{}) (*GitHubRelease, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &GitHubRelease{}, err
	}

	newHandler, err := github.New(github.Spec{
		Owner:      newSpec.Owner,
		Repository: newSpec.Repository,
		Token:      newSpec.Token,
		URL:        newSpec.URL,
		Username:   newSpec.Username,
	}, "")
	if err != nil {
		return &GitHubRelease{}, err
	}

	if !newSpec.KeepOriginalVersion && newSpec.VersionFilter.Kind == version.SEMVERVERSIONKIND {
		logrus.Warningf("%s\n\n", DeprecatedSemverVersionMessage)
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitHubRelease{}, err
	}

	return &GitHubRelease{
		ghHandler:           newHandler,
		versionFilter:       newFilter,
		keepOriginalVersion: newSpec.KeepOriginalVersion,
	}, nil
}
