package github

import (
	"context"
	"errors"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

// repositorySearchQuery defines the structure for the GraphQL search query
type repositorySearchQuery struct {
	RateLimit RateLimit
	Search    struct {
		RepositoryCount int
		PageInfo        PageInfo
		Edges           []struct {
			Cursor string
			Node   struct {
				// https://docs.github.com/en/graphql/reference/objects#repository
				Repository struct {
					//https://docs.github.com/en/graphql/reference/interfaces#repositoryowner
					Owner struct {
						Login string
					}
					Name          string
					NameWithOwner string
					IsArchived    bool
				} `graphql:"... on Repository"`
			}
		}
	} `graphql:"search(query: $query, type: REPOSITORY, first: $first, after: $after)"`
}

// SearchRepositories searches repositories based on a query string.
// It handles pagination and rate limiting.
// If the rate limit is exceeded, it will retry the request up to MaxRetry times.
// Query examples:
//   - "org:myorg" to list all repositories in an organization
//   - "user:myuser" to list all repositories for a user
//   - "topic:mytopic" to list all repositories with a specific topic
//   - "myrepo in:name" to search for repositories with "myrepo" in their name
//
// More examples can be found at https://github.com/search/advanced
func SearchRepositories(c client.Client, queryStr string, retry int, ctx context.Context) ([]string, error) {
	var allRepos []string
	var after githubv4.String

	for {
		query := repositorySearchQuery{}
		variables := map[string]interface{}{
			"query": githubv4.String(queryStr),
			"first": githubv4.Int(100),
			"after": after,
		}

		err := c.Query(ctx, &query, variables)
		if err != nil {
			if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) && retry < client.MaxRetry {
				rateLimit, err := queryRateLimit(c, ctx)
				if err != nil {
					logrus.Errorf("Error querying GitHub API rate limit: %s", err)
				}
				logrus.Debugln(rateLimit)
				if retry < client.MaxRetry {
					logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
					rateLimit.Pause()
					return SearchRepositories(c, queryStr, retry+1, ctx)
				}
				return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
			}
			return nil, err
		}

		for _, edge := range query.Search.Edges {
			allRepos = append(allRepos, edge.Node.Repository.NameWithOwner)
		}

		if query.Search.PageInfo.HasNextPage {
			endCursor := query.Search.PageInfo.EndCursor
			after = githubv4.String(endCursor)
		} else {
			break
		}
	}

	if len(allRepos) > 0 {
		return allRepos, nil
	}
	return nil, nil
}
