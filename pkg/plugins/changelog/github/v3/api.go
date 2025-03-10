package changelog

import (
	"context"
	"fmt"

	"github.com/google/go-github/v69/github"
	"github.com/sirupsen/logrus"
)

// getReleasesFromAPI returns a list of releases,
// which does not include regular Git tags that have not been associated with a release.
func getReleasesFromAPI(client *github.Client, owner, repository string) (allReleases []*github.RepositoryRelease, err error) {

	ctx := context.Background()

	opt := github.ListOptions{
		PerPage: 100,
	}

	for {
		releases, resp, err := client.Repositories.ListReleases(
			ctx,
			owner,
			repository,
			&opt)

		if err != nil {

			if _, ok := err.(*github.RateLimitError); ok {
				return nil, fmt.Errorf("GitHub rate limit reached")
			}

			// In addition to these rate limits, GitHub imposes a secondary rate limit on all API clients.
			// This rate limit prevents clients from making too many concurrent requests.
			if _, ok := err.(*github.AbuseRateLimitError); ok {
				return nil, fmt.Errorf("Secondary rate limit reached")
			}
			return nil, fmt.Errorf("listing GitHub releases: %w", err)
		}

		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			logrus.Debugf("GitHub Api limit: %q", resp.Rate.String())
			return allReleases, nil
		}

		opt.Page = resp.NextPage
	}
}
