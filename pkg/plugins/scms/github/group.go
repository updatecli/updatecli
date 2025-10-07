package github

import (
	"context"
	"errors"
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
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) && retry < MaxRetry {
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
				return getTeamID(client, org, team, retry+1)
			}
			return "", errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return "", err
	}

	logrus.Debugln(query.RateLimit)

	return query.Organization.Team.ID, nil
}
