package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/olblak/updateCli/pkg/plugins/version/semver"
	"github.com/sirupsen/logrus"

	"github.com/shurcooL/githubv4"
)

// Source retrieves a specific version tag from Github Releases.
func (g *Github) Source(workingDir string) (string, error) {

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
				releases(first:10, orderBy:$orderBy){
					nodes{
						name
						tagName
						isDraft
						isPrerelease
					}
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

	type release struct {
		Name         string
		TagName      string
		IsDraft      bool
		IsPrerelease bool
	}

	var query struct {
		Repository struct {
			Releases struct {
				Nodes []release
			} `graphql:"releases(first: 100, orderBy: $orderBy)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Owner),
		"repository": githubv4.String(g.Repository),
		"orderBy": githubv4.ReleaseOrder{
			Field:     "CREATED_AT",
			Direction: "DESC",
		},
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		logrus.Errorf("\u2717 Couldn't find a valid github release version")
		logrus.Errorf("\t %s", err)
		return "", err
	}

	value := ""
	versions := []string{}

	for _, release := range query.Repository.Releases.Nodes {
		if !release.IsDraft && !release.IsPrerelease {
			versions = append(versions, release.TagName)
		}
	}

	switch g.VersionType {
	case STRINGVERSIONTYPE:
		if g.Version == "latest" {
			value = versions[0]
		} else {
			for _, version := range versions {
				if strings.HasPrefix(version, g.Version) {
					value = version
					break
				}
			}
		}
	case SEMVERVERSIONTYPE:
		sv := semver.Semver{
			Constraint: g.Version,
		}
		err = sv.Init(versions)

		if err != nil {
			return value, err
		}

		value, err = sv.GetLatestVersion()
		if err != nil {
			return value, err
		}
	default:
		return value, fmt.Errorf("Something went wrong while decoding version %q for version pattern %q", g.Version, g.VersionType)
	}

	logrus.Debugf("%d version returned from Github", len(query.Repository.Releases.Nodes))
	logrus.Debugf("%s", versions)

	if len(value) == 0 {
		logrus.Infof("\u2717 No Github Release version founded matching pattern %q", g.Version)
	} else if len(value) > 0 {
		logrus.Infof("\u2714 Github Release version %q founded matching pattern %q", value, g.Version)
	} else {
		logrus.Errorf("Something unexpected happened in Github source")
	}

	return value, nil
}
