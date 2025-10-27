package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

// branchesQuery defines the structure for the GraphQL query to list branches
type branchesQuery struct {
	RateLimit  RateLimit
	Repository struct {
		Refs struct {
			Nodes []struct {
				Name string
			}
			PageInfo struct {
				HasNextPage bool
				EndCursor   githubv4.String
			}
		} `graphql:"refs(refPrefix: \"refs/heads/\", first: $first, after: $after)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// ListBranches lists all branches from a specific repository using pagination
func ListBranches(c client.Client, owner, repo string, retry int, ctx context.Context) ([]string, error) {
	var branches []string
	var after *githubv4.String

	for {
		var q branchesQuery
		vars := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(repo),
			"first": githubv4.Int(100),
			"after": after,
		}

		err := c.Query(ctx, &q, vars)
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
					return ListBranches(c, owner, repo, retry+1, ctx)
				}
				return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
			}
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}

		for _, node := range q.Repository.Refs.Nodes {
			branches = append(branches, node.Name)
		}

		if q.Repository.Refs.PageInfo.HasNextPage {
			after = &q.Repository.Refs.PageInfo.EndCursor
		} else {
			break
		}
	}

	return branches, nil
}
