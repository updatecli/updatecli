package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// query {
//   user(login: "USERNAME") {
//     id
//     name
//   }
// }

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

func getUserInfo(client GitHubClient, login string, retry int) (*userInfo, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(login),
	}

	var query userQuery

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(query.RateLimit)
			if retry < MaxRetry {
				query.RateLimit.Pause()

				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return getUserInfo(client, login, retry+1)
			}
			return nil, fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, fmt.Errorf("querying GitHub API: %w", err)
	}

	logrus.Debugln(query.RateLimit)

	return &query.User, nil
}
