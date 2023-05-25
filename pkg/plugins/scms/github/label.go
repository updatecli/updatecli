package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// labelsQuery defines a github v4 API query to retrieve a list of labels defined for the repository with their name and descriptions
/*
https://developer.github.com/v4/explorer/

query getLatestRelease{
	rateLimit {
		cost
		remaining
		resetAt
	}
	repository(owner: "updatecli", name: "updatecli"){
		labels (last: 5) {
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
type labelsQuery struct {
	RateLimit  RateLimit
	Repository struct {
		Labels repositoryLabels `graphql:"labels(last: 5, before: $before)"`
	} `graphql:"repository(owner: $owner, name: $repository)"`
}
type repositoryLabels struct {
	TotalCount int
	PageInfo   PageInfo
	Edges      []labelEdge
}
type labelEdge struct {
	Cursor string
	Node   labelNode
}
type labelNode struct {
	ID          string
	Name        string
	Description string
}

// repositoryLabelApi holds specific label information returned from GitHub API
type repositoryLabelApi struct {
	ID          string
	Name        string
	Description string
}

// getRepositoryLabels queries GitHub Api to retrieve every labels configured for a repository
func (g *Github) getRepositoryLabels() ([]repositoryLabelApi, error) {
	var repositoryLabels []repositoryLabelApi

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Spec.Owner),
		"repository": githubv4.String(g.Spec.Repository),
		"before":     (*githubv4.String)(nil),
	}

	var query labelsQuery

	for {
		err := g.client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return nil, err
		}

		query.RateLimit.Show()

		// Retrieve remote label information such as label ID, label name, labe description
		for _, node := range query.Repository.Labels.Edges {
			repositoryLabels = append(
				repositoryLabels,
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

	return repositoryLabels, nil
}

func mergeLabels(a, b []repositoryLabelApi) []repositoryLabelApi {

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
