package github

import (
	"context"
	"errors"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
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
func getTeamID(ctx context.Context, c client.Client, org string, team string, retry int) (string, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(org),
		"team":  githubv4.String(team),
	}

	var query getGroupQuery

	err := c.Query(ctx, &query, variables)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) && retry < client.MaxRetry {
			// If the query failed because we reached the rate limit,
			// then we need to re-requery the rate limit to get the latest information
			rateLimit, err := queryRateLimit(c, ctx)
			if err != nil {
				logrus.Errorf("Error querying GitHub API rate limit: %s", err)
			}
			logrus.Debugln(rateLimit)
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return getTeamID(ctx, c, org, team, retry+1)
			}
			return "", errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return "", err
	}

	logrus.Debugln(query.RateLimit)

	return query.Organization.Team.ID, nil
}
