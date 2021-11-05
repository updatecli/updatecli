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

// getRepositoryLabelsInformation queries GitHub Api to retrieve every labels configured for a repository
// then only return those matching label name specified via an updatecli configuration and those retrieve
// from a an open pull request
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
	var matchingLabels []repositoryLabel
	var repoLabels []repositoryLabel

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

		//// Merge remote pullrequest label and those provided via an updatecli spec
		//upToDateLabels := g.spec.PullRequest.Labels
		//for _, specLabel := range g.spec.PullRequest.Labels {
		//	found := false
		//	for _, remoteLabel := range g.remotePullRequest.Labels {
		//		if specLabel == remoteLabel {
		//			found = true
		//			break
		//		}
		//	}
		//	if !found {
		//		upToDateLabels = append(upToDateLabels, specLabel)
		//	}
		//}

		// Retrieve remote label information such as label ID, label name, labe description
		for _, node := range query.Repository.Labels.Edges {
			repoLabels = append(
				repoLabels,
				repositoryLabel{
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

	for _, l := range g.spec.PullRequest.Labels {
		for _, repoLabel := range repoLabels {

			if l == repoLabel.Name {
				matchingLabels = append(
					matchingLabels,
					repositoryLabel{
						ID:          repoLabel.ID,
						Name:        repoLabel.Name,
						Description: repoLabel.Description,
					})
				break
			}
		}
	}

	logrus.Debugf("%d labels found over %d requested", len(matchingLabels), len(g.spec.PullRequest.Labels))

	return matchingLabels, nil
}
