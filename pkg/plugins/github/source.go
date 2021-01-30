package github

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/shurcooL/githubv4"
)

// Source retrieves a specific version tag from Github Releases.
func (g *Github) Source(workingDir string) (string, error) {

	_, err := g.Check()
	if err != nil {
		return "", err
	}

	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getLatestRelease($owner: String!, $repository: String!){
			repository(owner: $owner, name: $repository){
				releases(first:10, orderBy:$orderBy){
					nodes{
						name
						tagName
						isDraft
						isPrerelease
					}
				}
			}
		}
		# Variables
		{
			"owner": "olblak",
			"repository": "charts",
		}
	*/

	client := g.NewClient()

	type release struct {
		Name         string
		TagName      string
		IsDraft      bool
		IsPrerelease bool
	}

	var query struct {
		Repository struct {
			Releases struct {
				Nodes []release
			} `graphql:"releases(first: 5, orderBy: $orderBy)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Owner),
		"repository": githubv4.String(g.Repository),
		"orderBy": githubv4.ReleaseOrder{
			Field:     "CREATED_AT",
			Direction: "DESC",
		},
	}

	err = client.Query(context.Background(), &query, variables)

	if err != nil {
		logrus.Errorf("\u2717 Couldn't find a valid github release version")
		logrus.Errorf("\t %s", err)
		return "", err
	}

	value := ""
	for _, release := range query.Repository.Releases.Nodes {
		if !release.IsDraft && !release.IsPrerelease {
			value = release.TagName
			break
		}
	}

	logrus.Infof("\u2714 '%s' github release version founded: %s", g.Version, value)
	return value, nil
}
