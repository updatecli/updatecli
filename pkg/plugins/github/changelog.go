package github

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/shurcooL/githubv4"
)

// Changelog contains various information used to describe target changes
type Changelog struct {
	Title       string
	Description string
	Report      string
}

// Changelog returns a changelog description based on a release name
func (g *Github) Changelog(name string) (string, error) {

	errs := g.Check()
	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorf("%s\n", e)
		}
		return "", fmt.Errorf("wrong github configuration")
	}

	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getLatestRelease($owner: String!, $repository: String!){
			repository(owner: $owner, name: $repository){
				release(tagName: 1){
					description
					publishedAt
					url
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

	var query struct {
		Repository struct {
			Release struct {
				Description string
				Url         string
				PublishedAt time.Time
			} `graphql:"release(tagName: $tagName)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Owner),
		"repository": githubv4.String(g.Repository),
		"tagName":    githubv4.String(name),
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		logrus.Warnf("\t %s", err)
		return "", err
	}

	changelog := ""

	if len(query.Repository.Release.Url) == 0 {
		changelog = fmt.Sprintf("No Github Release found for %s on https://github.com/%s/%s",
			name,
			g.Owner,
			g.Repository)
	} else {
		changelog = fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
			query.Repository.Release.PublishedAt.String(),
			query.Repository.Release.Url,
			query.Repository.Release.Description)

	}

	return changelog, nil
}
