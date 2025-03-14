package github

import (
	"context"

	"github.com/shurcooL/githubv4"
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

func getUserInfo(client GitHubClient, login string) (*userInfo, error) {

	variables := map[string]interface{}{
		"login": githubv4.String(login),
	}

	var query userQuery

	err := client.Query(context.Background(), &query, variables)

	query.RateLimit.Show()

	if err != nil {
		return nil, err
	}

	return &query.User, nil
}
