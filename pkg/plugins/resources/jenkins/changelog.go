package jenkins

import (
	"fmt"
)

// Changelog returns the link to the found Jenkins version's changelog
func (j Jenkins) Changelog() string {
	var changelogURI string
	switch j.spec.Release {
	case WEEKLY:
		changelogURI = "changelog"
	case STABLE:
		changelogURI = "changelog-stable"
	default:
		return ""
	}

	if j.foundVersion == "" {
		return ""
	}

	changelog := fmt.Sprintf(
		"Jenkins changelog is available at: https://www.jenkins.io/%s/#v%s\n",
		changelogURI,
		j.foundVersion,
	)

	return changelog
}
