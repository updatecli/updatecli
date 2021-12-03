package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// repositoryLabelApi holds specific label informations returned from Github API
type repositoryLabelApi struct {
	ID          string
	Name        string
	Description string
}

// getRepositoryLabelsInformation queries GitHub Api to retrieve every labels configured for a repository
// then only return those matching label name specified via an updatecli configuration
func (g *Github) GetRepositoryLabelsInformation() ([]repositoryLabelApi, error) {

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
	if len(g.Spec.PullRequest.Labels) == 0 {
		return nil, nil
	}
	var repoLabels []repositoryLabelApi

	client := g.NewClient()

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Spec.Owner),
		"repository": githubv4.String(g.Spec.Repository),
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

		// Retrieve remote label information such as label ID, label name, labe description
		for _, node := range query.Repository.Labels.Edges {
			repoLabels = append(
				repoLabels,
				repositoryLabelApi{
					ID:          node.Node.ID,
					Name:        node.Node.Name,
					Description: node.Node.Description,
				})
		}

		if !query.Repository.Labels.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.Labels.PageInfo.StartCursor))
	}

	matchingLabels := []repositoryLabelApi{}
	for _, l := range g.Spec.PullRequest.Labels {
		for _, repoLabel := range repoLabels {
			if l == repoLabel.Name {
				matchingLabels = append(matchingLabels, repoLabel)
			}
		}
	}

	return matchingLabels, nil
}

func MergeLabels(a, b []repositoryLabelApi) []repositoryLabelApi {

	result := b

	for i := 0; i < len(a); i++ {
		found := false
		for j := 0; j < len(b); j++ {
			if a[i].Name == b[j].Name {
				found = true
				break
			}
		}
		if !found {
			result = append(result, a[i])
		}
	}

	return result
}
