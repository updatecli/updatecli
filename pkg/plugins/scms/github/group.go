package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// getGroupQuery defines a github v4 API query to retrieve a group information
type getGroupQuery struct {
	RateLimit    RateLimit
	Organization struct {
		Team struct {
			ID string
		} `graphql:"team(slug: $team)"`
	} `graphql:"organization(login: $login)"`
}

// getTeamID return a group information from GitHub API
func getTeamID(client GitHubClient, org string, team string, retry int) (string, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(org),
		"team":  githubv4.String(team),
	}

	var query getGroupQuery

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) && retry < 3 {
			logrus.Debugln(query.RateLimit)
			if retry < MaxRetry {
				query.RateLimit.Pause()

				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return getTeamID(client, org, team, retry+1)
			}

			return "", fmt.Errorf("GitHub API rate limit exceeded. Please try again later")

		}

		return "", err
	}

	logrus.Debugln(query.RateLimit)

	return query.Organization.Team.ID, nil
}
