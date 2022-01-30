package githubrelease

import (
	"path"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/github"
	"github.com/updatecli/updatecli/pkg/plugins/version"
)

// GitHubRelease defines a resource of kind "githubrelease"
type GitHubRelease struct {
	ghHandler     github.GithubHandler
	versionFilter version.Filter
	foundVersion  version.Version
}

// New returns a new valid GitHubRelease object.
func New(spec interface{}) (*GitHubRelease, error) {

	newSpec := github.Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &GitHubRelease{}, err
	}

	if newSpec.Directory == "" {
		newSpec.Directory = path.Join(tmp.Directory, newSpec.Owner, newSpec.Repository)
	}

	if newSpec.URL == "" {
		newSpec.URL = "github.com"
	}

	newHandler, err := github.New(newSpec)
	if err != nil {
		return &GitHubRelease{}, err
	}

	return &GitHubRelease{
		ghHandler:     newHandler,
		versionFilter: newHandler.Spec.VersionFilter,
	}, nil
}
