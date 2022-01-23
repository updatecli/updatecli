package githubrelease

import "github.com/updatecli/updatecli/pkg/plugins/version"

func (gr *GitHubRelease) Changelog(version version.Version) (string, error) {
	return gr.ghHandler.Changelog(version)
}
