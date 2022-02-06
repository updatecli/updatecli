package github

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"

	"github.com/shurcooL/githubv4"
)

// Changelog contains various information used to describe target changes
type Changelog struct {
	Title       string
	Description string
	Report      string
}

// changelogQuery defines a github v4 API query to retrieve the changelog of a given release
/*
https://developer.github.com/v4/explorer/
# Query
query getLatestRelease($owner: String!, $repository: String!){
	repository(owner: $owner, name: $repository){
		release(tagName: "v0.17.0"){
			description
			publishedAt
			url
		}
	}
}
# Variables
{
	"owner": "updatecli",
	"repository": "updatecli"
}
*/
type changelogQuery struct {
	Repository struct {
		Release queriedRelease `graphql:"release(tagName: $tagName)"`
	} `graphql:"repository(owner: $owner, name: $repository)"`
}
type queriedRelease struct {
	Description string
	Url         string
	PublishedAt time.Time
}

// Changelog returns a changelog description based on a release name
func (g *Github) Changelog(version version.Version) (string, error) {

	// Github Release needs the original version, because the "found" version can be modified (semantic version without the prefix, transformed version, etc.)
	versionName := version.OriginalVersion

	var query changelogQuery

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Spec.Owner),
		"repository": githubv4.String(g.Spec.Repository),
		"tagName":    githubv4.String(versionName),
	}

	err := g.client.Query(context.Background(), &query, variables)
	if err != nil {
		logrus.Warnf("\t %s", err)
		return "", err
	}

	if len(query.Repository.Release.Url) == 0 {
		// TODO: getRepositoryURL()
		return fmt.Sprintf("No Github Release found for %s on https://github.com/%s/%s",
			versionName,
			g.Spec.Owner,
			g.Spec.Repository), nil
	}

	return fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
		query.Repository.Release.PublishedAt.String(),
		query.Repository.Release.Url,
		query.Repository.Release.Description), nil
}
