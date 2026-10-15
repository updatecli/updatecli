package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
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

// defaultLabelColor is the color assigned to new labels
const defaultLabelColor = "ededed"

// createLabel creates a new label on the repository identified by repositoryID
func (g *Github) createLabel(ctx context.Context, repositoryID, name, color string, retry int) (repositoryLabelApi, error) {
	var mutation struct {
		CreateLabel struct {
			Label labelNode
		} `graphql:"createLabel(input: $input)"`
	}

	if color == "" {
		color = defaultLabelColor
	}

	input := githubv4.CreateLabelInput{
		RepositoryID: githubv4.ID(repositoryID),
		Name:         githubv4.String(name),
		Color:        githubv4.String(color),
	}

	err := g.client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			rateLimit, rlErr := queryRateLimit(g.client, ctx)
			if rlErr != nil {
				logrus.Errorf("Error querying GitHub API rate limit: %s", rlErr)
			}

			logrus.Debugln(rateLimit)
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return g.createLabel(ctx, repositoryID, name, color, retry+1)
			}
			return repositoryLabelApi{}, errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return repositoryLabelApi{}, fmt.Errorf("creating label %q: %w", name, err)
	}

	return repositoryLabelApi{
		ID:          mutation.CreateLabel.Label.ID,
		Name:        mutation.CreateLabel.Label.Name,
		Description: mutation.CreateLabel.Label.Description,
	}, nil
}

// getRepositoryLabels queries GitHub Api to retrieve every labels configured for a repository
func (g *Github) getRepositoryLabels(ctx context.Context, retry int) ([]repositoryLabelApi, error) {
	var repositoryLabels []repositoryLabelApi

	variables := map[string]interface{}{
		keyOwner:      githubv4.String(g.Spec.Owner),
		keyRepository: githubv4.String(g.Spec.Repository),
		keyBefore:     (*githubv4.String)(nil),
	}

	var query labelsQuery

	for {
		err := g.client.Query(ctx, &query, variables)

		if err != nil {
			if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
				rateLimit, err := queryRateLimit(g.client, ctx)
				if err != nil {
					logrus.Errorf("Error querying GitHub API rate limit: %s", err)
				}

				logrus.Debugln(rateLimit)
				if retry < client.MaxRetry {
					logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
					rateLimit.Pause()
					return g.getRepositoryLabels(ctx, retry+1)
				}
				return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
			}
			return nil, fmt.Errorf("querying GitHub API: %w", err)
		}

		logrus.Debugln(query.RateLimit)

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

		variables[keyBefore] = githubv4.NewString(githubv4.String(query.Repository.Labels.PageInfo.StartCursor))
	}

	return repositoryLabels, nil
}

// matchLabels splits wanted label names into the ones already in the repository and the ones that aren't
func matchLabels(wanted []string, repositoryLabels []repositoryLabelApi) (matched []repositoryLabelApi, missing []string) {
	for _, name := range wanted {
		found := false
		for _, repoLabel := range repositoryLabels {
			if name == repoLabel.Name {
				matched = append(matched, repoLabel)
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, name)
		}
	}

	return matched, missing
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
