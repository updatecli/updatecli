package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// userQuery defines a github v4 API query to retrieve a user information
type userQuery struct {
	RateLimit RateLimit
	User      userInfo `graphql:"user(login: $login)"`
}

// userInfo defines a user information returned from GitHub API
type userInfo struct {
	ID   string
	Name string
}

// getUserInfo return a user information from GitHub API
func getUserInfo(client GitHubClient, login string, retry int) (*userInfo, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(login),
	}

	var query userQuery

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			// If the query failed because we reached the rate limit,
			// then we need to re-requery the rate limit to get the latest information
			rateLimit, err := queryRateLimit(client, context.Background())
			if err != nil {
				logrus.Errorf("Error querying GitHub API rate limit: %s", err)
			}

			logrus.Debugln(rateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				rateLimit.Pause()
				return getUserInfo(client, login, retry+1)
			}
			return nil, fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, fmt.Errorf("querying GitHub API: %w", err)
	}

	logrus.Debugln(query.RateLimit)

	return &query.User, nil
}
