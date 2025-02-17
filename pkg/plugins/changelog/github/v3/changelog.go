package changelog

import (
	"fmt"
	"os"

	"github.com/google/go-github/v69/github"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/changelog"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Changelog struct {
	URL           string
	Owner         string
	Repository    string
	Token         string
	VersionFilter version.Filter
}

// Search returns a list of changelogs, retrieved from a GitHub api, between two versions
func (c *Changelog) Search(from, to string) ([]changelog.Changelog, error) {

	var err error

	client := github.NewClient(nil)

	token := c.Token

	if token == "" {
		if os.Getenv("GITHUB_TOKEN") != "" {
			token = os.Getenv("GITHUB_TOKEN")
		} else if os.Getenv("UPDATECLI_GITHUB_TOKEN") != "" {
			token = os.Getenv("UPDATECLI_GITHUB_TOKEN")
		}
	}

	if token != "" {
		client = client.WithAuthToken(token)
	}

	if c.URL != "" {
		client, err = client.WithEnterpriseURLs(c.URL, c.URL)
		if err != nil {
			return nil, fmt.Errorf("configure enterprise url: %w", err)
		}
	}

	releasesID := generateCatalogID(c.URL, c.Owner, c.Repository, from, to)

	allReleases := getReleasesFromCatalog(releasesID)

	if allReleases == nil {
		logrus.Debugf("Changelog releases not detected locally, checking online")

		allReleases, err = getReleasesFromAPI(client, c.Owner, c.Repository)
		if err != nil {
			return nil, fmt.Errorf("fetching GitHub releases: %w", err)
		}
	}

	switch c.VersionFilter.Kind {
	case version.SEMVERVERSIONKIND:
		sortReleasesBySemver(&allReleases)
	case "":
		if isSemverDetected(from, to) {
			sortReleasesBySemver(&allReleases)
		}
	default:
		logrus.Debugf("version filter of kind %q not supported. Feel free to open an issue explaining your need", c.VersionFilter.Kind)
	}

	allReleases = filterReleases(allReleases, from, to)

	if allReleases != nil {
		if Catalog == nil {
			Catalog = make(map[string][]*github.RepositoryRelease)
		}
		Catalog[releasesID] = allReleases
	}

	return convertToChangelog(allReleases), nil
}

// filterReleases filters releases between two versions
func filterReleases(allReleases []*github.RepositoryRelease, from, to string) []*github.RepositoryRelease {

	if from == "" && to == "" {
		return allReleases
	}

	var filteredReleases []*github.RepositoryRelease

	foundFrom := false
	foundTo := false

	for _, release := range allReleases {

		if from != "" {
			if release.GetTagName() == from {
				foundFrom = true
			}
		}

		if foundFrom {
			filteredReleases = append(filteredReleases, release)
		}

		if to != "" {
			if release.GetTagName() == to {
				foundTo = true
				break
			}
		}
	}

	if len(filteredReleases) == 0 {
		return nil
	}

	if to != "" && !foundTo {
		logrus.Warnf("Release %q not found so I only return the found release", to)
		return filteredReleases[0:1]
	}

	return filteredReleases
}

// convertToChangelog converts a list of github.RepositoryRelease to a list of changelog.Changelog
// so we can use it from Updatecli
func convertToChangelog(releases []*github.RepositoryRelease) []changelog.Changelog {
	var changelogs []changelog.Changelog

	for _, release := range releases {
		changelog := changelog.Changelog{
			Title:       release.GetTagName(),
			Body:        release.GetBody(),
			PublishedAt: release.GetPublishedAt().String(),
			URL:         *release.HTMLURL,
		}
		changelogs = append(changelogs, changelog)
	}

	return changelogs
}
