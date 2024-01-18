package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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
	g.mu.RLock()
	defer g.mu.RUnlock()

	// GitHub Release needs the original version, because the "found" version can be modified (semantic version without the prefix, transformed version, etc.)
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

	URL, err := url.JoinPath(g.Spec.URL, g.Spec.Owner, g.Spec.Repository)

	if err != nil {
		return "", err
	}

	if len(query.Repository.Release.Url) == 0 {
		// TODO: getRepositoryURL()
		return fmt.Sprintf("no GitHub Release found for %s on %q", versionName, URL), nil
	}

	return fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
		query.Repository.Release.PublishedAt.String(),
		query.Repository.Release.Url,
		query.Repository.Release.Description), nil
}

// ChangelogV3 returns a changelog description based on a release name using the GitHub api v3 version
func (g *Github) ChangelogV3(version string) (string, error) {
	URL := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s",
		g.Spec.URL, g.Spec.Owner, g.Spec.Repository, version)

	logrus.Debugf("Retrieving changelog from %q", URL)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		logrus.Debugf("failed to retrieve changelog from GitHub %v\n", err)
		return "", err
	}

	// Guard the getTokenFromEnv() function call with a mutex to ensure that the environment variables are accessed safely
	g.mu.Lock()
	envToken := getTokenFromEnv()
	g.mu.Unlock()

	if envToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", envToken))
		req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Debugf("failed to retrieve changelog from GitHub %v\n", err)
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		if err != nil {
			logrus.Debugf("failed to retrieve changelog from GitHub %v\n", err)
		}
		logrus.Debugf("\n%v\n", string(body))
		return "", err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting GitHub api data %v\n", err)
		return "", err
	}

	type ReleaseInfo struct {
		HtmlURL string `json:"html_url,"`
		Body    string `json:"body,"`
	}

	release := ReleaseInfo{}

	err = json.Unmarshal(data, &release)
	if err != nil {
		logrus.Errorf("error unmarshalling json: %v", err)
		return "", err
	}

	return fmt.Sprintf("Changelog retrieved from:\n\t%s\n%s",
		release.HtmlURL, release.Body), nil
}

func getTokenFromEnv() string {
	// Lock the mutex to prevent concurrent access to the function
	g := Github{}
	g.mu.Lock()
	defer g.mu.Unlock()

	if envToken := os.Getenv("UPDATECLI_GITHUB_TOKEN"); envToken != "" {
		logrus.Debugln("GitHub token defined by environment variable UPDATECLI_GITHUB_TOKEN detected")
		return envToken
	}

	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		logrus.Debugln("GitHub token defined by environment variable GITHUB_TOKEN detected")
		return envToken
	}

	return ""
}
