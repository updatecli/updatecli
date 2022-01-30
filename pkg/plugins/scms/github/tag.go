package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// tagsQuery defines a github v4 API query to retrieve a list of tags sorted in reverse order of commit tags.
/*
https://developer.github.com/v4/explorer/
# Query
query getLatestTag($cursor: String){
	rateLimit {
		cost
		remaining
		resetAt
	}
	repository(owner: "kubernetes", name: "kubectl") {
		refs(refPrefix: "refs/tags/", first: 100, after: $cursor, orderBy: {field: TAG_COMMIT_DATE, direction: DESC}) {
			totalCount
			pageInfo {
				hasNextPage
				endCursor
			}
			edges {
				node {
						name
				}
				cursor
			}
		}
	}
}
*/
type tagsQuery struct {
	RateLimit  RateLimit
	Repository struct {
		Refs repositoryRef `graphql:"refs(refPrefix: $refPrefix, last: 100, before: $before,orderBy: $orderBy)"`
	} `graphql:"repository(owner: $owner, name: $repository)"`
}
type repositoryRef struct {
	TotalCount int
	PageInfo   PageInfo
	Edges      []refEdge
}
type refNode struct {
	Name string
}
type refEdge struct {
	Cursor string
	Node   refNode
}

// SearchTags return every tags from the github api return in reverse order of commit tags.
func (g *Github) SearchTags() (tags []string, err error) {
	var query tagsQuery

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Spec.Owner),
		"repository": githubv4.String(g.Spec.Repository),
		"refPrefix":  githubv4.String("refs/tags/"),
		"before":     (*githubv4.String)(nil),
		"orderBy": githubv4.RefOrder{
			Field:     "TAG_COMMIT_DATE",
			Direction: "DESC",
		},
	}

	expectedFound := 0
	tagCounter := 0
	for {
		err = g.client.Query(context.Background(), &query, variables)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		expectedFound = query.Repository.Refs.TotalCount

		query.RateLimit.Show()

		for i := len(query.Repository.Refs.Edges) - 1; i >= 0; i-- {
			tagCounter++
			node := query.Repository.Refs.Edges[i]
			tags = append(tags, node.Node.Name)
		}

		if !query.Repository.Refs.PageInfo.HasPreviousPage {
			break
		}
		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.Refs.PageInfo.StartCursor))
	}

	if expectedFound != tagCounter {
		return tags, fmt.Errorf("something went wrong, found %d, expected %d", tagCounter, expectedFound)
	}

	logrus.Debugf("%d tags found", len(tags))

	return tags, err
}
