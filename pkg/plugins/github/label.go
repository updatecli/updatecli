package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// repositoryLabel hold specific label information returned from Github API
type repositoryLabel struct {
	ID          string
	Name        string
	Description string
}

// getRepositoryLabelsInformation query GitHub Api to retrieve every labels configured for a repository
// then only return those matching label name specified via an updatecli configuration
func (g *Github) getRepositoryLabelsInformation() ([]repositoryLabel, error) {

	/*
		https://developer.github.com/v4/explorer/

			query($owner: String!, $name: String!) {
				rateLimit {
					cost
					remaining
					resetAt
				}
				repository(owner: $owner, name: $name){
					labels (last: 5, before: $before) {
						totalCount
						pageInfo {
							hasNextPage
							endCursor
						}
						edges {
							node {
								id
								name
								description
							}
							cursor
						}
					}
				}
			}

	*/

	// Early exit as no label information are needed
	if len(g.spec.PullRequest.Labels) == 0 {
		return nil, nil
	}
	var labels []repositoryLabel

	client := g.NewClient()

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.spec.Owner),
		"repository": githubv4.String(g.spec.Repository),
		"before":     (*githubv4.String)(nil),
	}

	var query struct {
		RateLimit  RateLimit
		Repository struct {
			Labels struct {
				TotalCount int
				PageInfo   PageInfo
				Edges      []struct {
					Cursor string
					Node   struct {
						ID          string
						Name        string
						Description string
					}
				}
			} `graphql:"labels(last: 5, before: $before)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	for {
		err := client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return nil, err
		}

		query.RateLimit.Show()

		for _, l := range g.spec.PullRequest.Labels {
			found := false
			for _, node := range query.Repository.Labels.Edges {

				if l == node.Node.Name {
					found = true
					labels = append(
						labels,
						repositoryLabel{
							ID:          node.Node.ID,
							Name:        node.Node.Name,
							Description: node.Node.Description,
						})
					break
				}
			}
			if !found {
				logrus.Debugf("Label %q not defined on repository %s/%s, ignoring it", l, g.spec.Owner, g.spec.Repository)
			}

		}

		if !query.Repository.Labels.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.Labels.PageInfo.StartCursor))
	}

	logrus.Debugf("%d labels found over %d requested", len(labels), len(g.spec.PullRequest.Labels))

	return labels, nil
}
