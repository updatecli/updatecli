package jenkins

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Changelog returns the link to the found Jenkins version's changelog
func (j Jenkins) Changelog(from, to string) *result.Changelogs {

	if from == "" && to == "" {
		return nil
	}

	var changelogURI string
	switch j.spec.Release {
	case WEEKLY:
		changelogURI = "changelog"
	case STABLE:
		changelogURI = "changelog-stable"
	default:
		return nil
	}

	changelog := fmt.Sprintf(
		"Jenkins changelog is available at: https://www.jenkins.io/%s/#v%s\n",
		changelogURI,
		from,
	)

	url := fmt.Sprintf("https://www.jenkins.io/%s/#v%s", changelogURI, from)

	return &result.Changelogs{
		{
			Title: from,
			Body:  changelog,
			URL:   url,
		},
	}
}
