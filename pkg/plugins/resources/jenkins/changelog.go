package jenkins

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	weeklyRegex *regexp.Regexp = regexp.MustCompile(`\d+\.\d+`)
	stableRegex *regexp.Regexp = regexp.MustCompile(`\d+\.\d+\.\d+`)
)

// Changelog returns the link to the found Jenkins version's changelog
func (j Jenkins) Changelog(from, to string) *result.Changelogs {

	_, versions, err := j.getVersions()
	if err != nil {
		logrus.Debugf("ignored error, searching changelogs: %s", err)
		return nil
	}

	releases := filterReleases(versions, j.spec.Release, from, to)

	if len(releases) == 0 {
		logrus.Debugf("No changelog found")
		return nil
	}

	var changelogs result.Changelogs

	for i := range releases {

		changelog := result.Changelog{
			Title: releases[i],
		}

		switch j.spec.Release {
		case WEEKLY:
			changelog.URL = fmt.Sprintf("https://www.jenkins.io/changelog/#v%s", releases[i])
			changelog.Body = fmt.Sprintf("Jenkins changelog is available at: https://www.jenkins.io/changelog/#v%s\n", releases[i])
		case STABLE:
			changelog.URL = fmt.Sprintf("https://www.jenkins.io/changelog-stable/#v%s", releases[i])
			changelog.Body = fmt.Sprintf("Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v%s\n", releases[i])
		}

		changelogs = append(changelogs, changelog)
	}

	return &changelogs

}

func filterReleases(allReleases []string, releaseName, from, to string) []string {
	if from == "" && to == "" {
		return allReleases
	}

	var filteredReleases []string

	foundFrom := false
	foundTo := false

	for _, release := range allReleases {

		switch releaseName {
		case WEEKLY:
			if !weeklyRegex.MatchString(release) {
				continue
			}
		case STABLE:
			if !stableRegex.MatchString(release) {
				continue
			}
		}

		if from != "" {
			if release == from {
				foundFrom = true
			}
		}

		if to != "" {
			if release == to {
				filteredReleases = append(filteredReleases, release)
				foundTo = true
				break
			}
		}

		if foundFrom {
			filteredReleases = append(filteredReleases, release)
		}
	}

	if len(filteredReleases) == 0 {
		return nil
	}

	if from != "" && !foundTo {
		logrus.Debugf("Jenkins release version %q not found so I only return the latest release", from)
		return filteredReleases[0:1]
	}

	slices.Reverse(filteredReleases)

	return filteredReleases
}
