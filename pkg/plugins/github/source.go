package github

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/shurcooL/githubv4"
)

// Source retrieves a specific version tag from Github Releases.
func (g *Github) Source(workingDir string) (value string, err error) {

	versions, err := g.SearchReleases()

	if err != nil {
		logrus.Error(err)
		return value, err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No GitHub Release found. As fallback Looking at published git tags", result.ATTENTION)
		versions, err = g.SearchTags()
		if err != nil {
			logrus.Errorf("%s", err)
			return "", err
		}
		if len(versions) == 0 {
			logrus.Infof("\t=> No release or git tags found, exiting")
			return "", fmt.Errorf("No release or git tags found, exiting")
		}
	}

	err = g.spec.VersionFilter.Search(versions)
	if err != nil {
		return "", err
	}
	g.foundVersion = g.spec.VersionFilter.FoundVersion
	value = g.foundVersion.ParsedVersion

	if len(value) == 0 {
		logrus.Infof("%s No Github Release version found matching pattern %q", result.FAILURE, g.spec.VersionFilter.Pattern)
		return value, fmt.Errorf("no Github Release version found matching pattern %q", g.spec.VersionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("%s Github Release version %q found matching pattern %q", result.SUCCESS, value, g.spec.VersionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in Github source")
	}

	return value, nil
}

// SearchTags return every tags from the github api return in reverse order of commit tags.
func (g *Github) SearchTags() (tags []string, err error) {

	client := g.NewClient()

	//	  rateLimit {
	//	    cost
	//	    remaining
	//	    resetAt
	//	  }
	//	  repository(owner: "kubernetes", name: "kubectl") {
	//	    refs(refPrefix: "refs/tags/", first: 100, after: $cursor, orderBy: {field: TAG_COMMIT_DATE, direction: DESC}) {
	//	      totalCount
	//	      pageInfo {
	//	        hasNextPage
	//	        endCursor
	//	      }
	//	      edges {
	//	        node {
	//	            name
	//	        }
	//	        cursor
	//	      }
	//	    }
	//	  }
	//	}

	var query struct {
		RateLimit  RateLimit
		Repository struct {
			Refs struct {
				TotalCount int
				PageInfo   PageInfo
				Edges      []struct {
					Cursor string
					Node   struct {
						Name string
					}
				}
			} `graphql:"refs(refPrefix: $refPrefix, last: 100, before: $before,orderBy: $orderBy)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.spec.Owner),
		"repository": githubv4.String(g.spec.Repository),
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
		err = client.Query(context.Background(), &query, variables)
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
		return tags, fmt.Errorf("Something went wrong, find %d, expected %d", tagCounter, expectedFound)
	}

	logrus.Debugf("%d tags found", len(tags))

	return tags, err
}

// SearchReleases return every releases from the github api returned in reverse order of created time.
func (g *Github) SearchReleases() (releases []string, err error) {
	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getLatestRelease($owner: String!, $repository: String!){
			rateLimit {
				cost
				remaining
				resetAt
			}
			repository(owner: $owner, name: $repository){
				releases(last:100, before: $before, orderBy:$orderBy){
		    		totalCount
		    		pageInfo {
		    		  hasNextPage
		    		  endCursor
		    		}
		    		edges {
		    		  node {
		    		      	name
				  			tagName
							isDraft
							isPrerelease
		    		  }
		    		  cursor
		    		}
				}
			}
		}
		# Variables
		{
			"owner": "olblak",
			"repository": "charts"
		}
	*/

	client := g.NewClient()

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.spec.Owner),
		"repository": githubv4.String(g.spec.Repository),
		"before":     (*githubv4.String)(nil),
		"orderBy": githubv4.ReleaseOrder{
			Field:     "CREATED_AT",
			Direction: "DESC",
		},
	}

	var query struct {
		RateLimit  RateLimit
		Repository struct {
			Releases struct {
				TotalCount int
				PageInfo   PageInfo
				Edges      []struct {
					Cursor string
					Node   struct {
						Name         string
						TagName      string
						IsDraft      bool
						IsPrerelease bool
					}
				}
			} `graphql:"releases(last: 100, before: $before, orderBy: $orderBy)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	expectedFound := 0
	releaseCounter := 0

	for {
		err := client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return releases, err
		}

		query.RateLimit.Show()

		for i := len(query.Repository.Releases.Edges) - 1; i >= 0; i-- {
			releaseCounter++
			node := query.Repository.Releases.Edges[i]
			if !node.Node.IsDraft && !node.Node.IsPrerelease {
				releases = append(releases, node.Node.TagName)
			}
		}

		expectedFound = query.Repository.Releases.TotalCount

		if !query.Repository.Releases.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.Releases.PageInfo.StartCursor))
	}

	if expectedFound != releaseCounter {
		return releases, fmt.Errorf("Something went wrong, found %d releases, expected %d", releaseCounter, expectedFound)
	}

	logrus.Debugf("%d releases found", len(releases))
	return releases, nil

}
