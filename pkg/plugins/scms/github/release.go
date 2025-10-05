package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// releasesQuery defines a github v4 API query to retrieve a list of releases sorted by reverse order of created time.
/*
https://developer.github.com/v4/explorer/
# Query
query getLatestRelease($owner: String!, $repository: String!, $before: String, $orderBy: ReleaseOrder) {
  rateLimit {
    cost
    remaining
    resetAt
  }
  repository(owner: $owner, name: $repository) {
    releases(last: 100, before: $before, orderBy: $orderBy) {
      totalCount
      pageInfo {
        hasNextPage
        endCursor
      }
      edges {
        node {
          name
          tagName
          tagCommit {
            oid
          }
          isDraft
          isPrerelease
        }
        cursor
      }
    }
  }
}# Variables
{
  "owner": "updatecli",
  "repository": "updatecli",
  "before": null,
  "orderBy": {
    "field": "CREATED_AT",
    "direction": "DESC"
  }
}*/
// releasesQuery defines a github v4 API query to retrieve a list of releases sorted by reverse order of created time.
type releasesQuery struct {
	RateLimit  RateLimit
	Repository struct {
		Releases repositoryRelease `graphql:"releases(last: 100, before: $before, orderBy: $orderBy)"`
	} `graphql:"repository(owner: $owner, name: $repository)"`
}
type ReleaseNode struct {
	Name         string
	TagName      string
	TagCommit    TagCommit
	IsDraft      bool
	IsLatest     bool
	IsPrerelease bool
}
type TagCommit struct {
	Oid string
}
type releaseEdge struct {
	Cursor string
	Node   ReleaseNode
}
type repositoryRelease struct {
	TotalCount int
	PageInfo   PageInfo
	Edges      []releaseEdge
}

// SearchReleases return every releaseNode from the github api
// ordered by reverse order of created time.
// Draft and pre-releases are filtered out.
func (g *Github) SearchReleases(releaseType ReleaseType, retry int) (releases []ReleaseNode, err error) {
	var query releasesQuery

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Spec.Owner),
		"repository": githubv4.String(g.Spec.Repository),
		"before":     (*githubv4.String)(nil),
		"orderBy": githubv4.ReleaseOrder{
			Field:     "CREATED_AT",
			Direction: "DESC",
		},
	}

	for {
		err := g.client.Query(context.Background(), &query, variables)
		if err != nil {
			if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
				logrus.Debugln(query.RateLimit)
				if retry < MaxRetry {
					logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
					query.RateLimit.Pause()
					return g.SearchReleases(releaseType, retry+1)
				}
				return nil, fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
			}
			return nil, fmt.Errorf("querying GitHub API: %w", err)
		}

		logrus.Debugln(query.RateLimit)

		for i := len(query.Repository.Releases.Edges) - 1; i >= 0; i-- {
			node := query.Repository.Releases.Edges[i]

			// If releaseType.Latest is set to true, then it means
			// we only care about identifying the latest release
			if releaseType.Latest {
				if node.Node.IsLatest {
					releases = append(releases, node.Node)
					break
				}
				// Check if the next release is of type "latest"
				continue
			}

			if node.Node.IsDraft {
				if releaseType.Draft {
					releases = append(releases, node.Node)
				}
			} else if node.Node.IsPrerelease {
				if releaseType.PreRelease {
					releases = append(releases, node.Node)
				}
			} else {
				if releaseType.Release {
					releases = append(releases, node.Node)
				}
			}
		}

		if !query.Repository.Releases.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.Releases.PageInfo.StartCursor))
	}

	logrus.Debugf("%d releases found", len(releases))
	return releases, nil
}

// SearchReleasesByTagName return every releases tag name from the github api
// ordered by reverse order of created time.
// Draft and pre-releases are filtered out.
func (g *Github) SearchReleasesByTagName(releaseType ReleaseType) (releases []string, err error) {
	releaseNodes, err := g.SearchReleases(releaseType, 0)
	if err != nil {
		logrus.Errorf("\t%s", err)
		return releases, err
	}

	for _, release := range releaseNodes {
		releases = append(releases, release.TagName)
	}
	return releases, nil
}

// SearchReleasesByTagHash return every releases tag hash from the github api
// ordered by reverse order of created time.
// Draft and pre-releases are filtered out.
func (g *Github) SearchReleasesByTagHash(releaseType ReleaseType) (releases []string, err error) {
	releaseNodes, err := g.SearchReleases(releaseType, 0)
	if err != nil {
		logrus.Errorf("\t%s", err)
		return releases, err
	}

	for _, release := range releaseNodes {
		releases = append(releases, release.TagCommit.Oid)
	}
	return releases, nil
}

// SearchReleasesByTitle return every releases title from the github api
// ordered by reverse order of created time.
// Draft and pre-releases are filtered out.
func (g *Github) SearchReleasesByTitle(releaseType ReleaseType) (releases []string, err error) {
	releaseNodes, err := g.SearchReleases(releaseType, 0)
	if err != nil {
		logrus.Errorf("\t%s", err)
		return releases, err
	}

	for _, release := range releaseNodes {
		releases = append(releases, release.Name)
	}
	return releases, nil
}
