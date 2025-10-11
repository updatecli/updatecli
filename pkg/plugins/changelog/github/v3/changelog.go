package changelog

import (
	"fmt"

	"github.com/google/go-github/v69/github"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
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
func (c *Changelog) Search(from, to string) (result.Changelogs, error) {

	var err error

	client := github.NewClient(nil)

	// No need to retrieve token from environment variable as
	// normally this should be already done from a different place
	// in the code and provided to this struct.
	token := c.Token

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

		if to != "" {
			if release.GetTagName() == to {
				foundTo = true
			}
		}

		if from != "" {
			if release.GetTagName() == from {
				filteredReleases = append(filteredReleases, release)
				foundFrom = true
				break
			}
		}

		if foundTo {
			filteredReleases = append(filteredReleases, release)
		}
	}

	if len(filteredReleases) == 0 {
		return nil
	}

	if from != "" && !foundFrom {
		logrus.Debugf("GitHub release version %q not found so I only return the latest release", from)
		return filteredReleases[0:1]
	}

	return filteredReleases
}

// convertToChangelog converts a list of github.RepositoryRelease to a list of result.Changelog
// so we can use it from Updatecli
func convertToChangelog(releases []*github.RepositoryRelease) []result.Changelog {
	var changelogs result.Changelogs

	for _, release := range releases {
		changelog := result.Changelog{
			Title:       release.GetTagName(),
			Body:        release.GetBody(),
			PublishedAt: release.GetPublishedAt().String(),
			URL:         *release.HTMLURL,
		}
		changelogs = append(changelogs, changelog)
	}

	return changelogs
}
