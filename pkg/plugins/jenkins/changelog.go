package jenkins

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/plugins/version"
)

// Changelog returns a changelog description based on a Jenkins version
// retrieved from a GitHub Release
func (j *Jenkins) Changelog(release version.Version) (string, error) {
	var changelogURI string
	switch j.spec.Release {
	case WEEKLY:
		changelogURI = "changelog"
	case STABLE:
		changelogURI = "changelog-stable"
	default:
		return "", fmt.Errorf("Unknown Jenkins release type: %q", j.spec.Release)
	}

	if release.ParsedVersion == "" {
		return "", fmt.Errorf("Empty Jenkins version: %q", release.ParsedVersion)
	}

	changelog := fmt.Sprintf(
		"Jenkins changelog is available at: https://www.jenkins.io/%s/#v%s\n",
		changelogURI,
		release.ParsedVersion,
	)

	return changelog, nil
}
