package github

import (
	"context"
	"errors"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
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
func getUserInfo(c client.Client, login string, retry int) (*userInfo, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(login),
	}

	var query userQuery

	err := c.Query(context.Background(), &query, variables)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			// If the query failed because we reached the rate limit,
			// then we need to re-requery the rate limit to get the latest information
			rateLimit, err := queryRateLimit(c, context.Background())
			if err != nil {
				logrus.Errorf("Error querying GitHub API rate limit: %s", err)
			}

			logrus.Debugln(rateLimit)
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return getUserInfo(c, login, retry+1)
			}
			return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, err
	}

	logrus.Debugln(query.RateLimit)

	return &query.User, nil
}
