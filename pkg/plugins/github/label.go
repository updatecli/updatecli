package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// repositoryLabel holds specific label informations returned from Github API
type repositoryLabel struct {
	ID          string
	Name        string
	Description string
}

// getRepositoryLabelsInformation queries GitHub Api to retrieve every labels configured for a repository
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

	matchingLabels := []repositoryLabel{}
	for _, l := range g.spec.PullRequest.Labels {
		for _, repoLabel := range repoLabels {
			if l == repoLabel.Name {
				matchingLabels = append(matchingLabels, repoLabel)
			}
		}
	}

	return matchingLabels, nil
}

// getPullRequestLabelsInformation queries GitHub Api to retrieve every labels assigned to a pullRequest
func (g *Github) getPullRequestLabelsInformation() ([]repositoryLabel, error) {

	/*
		query getPullRequests(
			$owner: String!,
			$name:String!,
			$before:Int!){
				repository(owner: $owner, name: $name) {
					pullRequest(number: 4){
		            labels(last: 5, before:$before){
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
					}
	*/

	client := g.NewClient()

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.spec.Owner),
		"repository": githubv4.String(g.spec.Repository),
		"number":     githubv4.Int(g.remotePullRequest.Number),
		"before":     (*githubv4.String)(nil),
	}

	var query struct {
		RateLimit  RateLimit
		Repository struct {
			PullRequest struct {
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
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	var pullRequestLabels []repositoryLabel
	for {
		err := client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return nil, err
		}

		query.RateLimit.Show()

		// Retrieve remote label information such as label ID, label name, labe description
		for _, node := range query.Repository.PullRequest.Labels.Edges {
			pullRequestLabels = append(
				pullRequestLabels,
				repositoryLabel{
					ID:          node.Node.ID,
					Name:        node.Node.Name,
					Description: node.Node.Description,
				})
		}

		if !query.Repository.PullRequest.Labels.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.PullRequest.Labels.PageInfo.StartCursor))
	}
	return pullRequestLabels, nil
}

func mergeLabels(a, b []repositoryLabel) []repositoryLabel {

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
