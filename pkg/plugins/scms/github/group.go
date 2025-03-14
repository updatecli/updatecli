package github

import (
	"context"

	"github.com/shurcooL/githubv4"
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
func getTeamID(client GitHubClient, org string, team string) (string, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(org),
		"team":  githubv4.String(team),
	}

	var query getGroupQuery

	err := client.Query(context.Background(), &query, variables)

	query.RateLimit.Show()

	if err != nil {
		return "", err
	}

	return query.Organization.Team.ID, nil
}
