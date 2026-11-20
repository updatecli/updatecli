package npm

import (
	"context"
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"

	giteaChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/gitea"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
	"github.com/updatecli/updatecli/pkg/plugins/changelog/vprefix"
)

// Changelog returns the link to the found npm package version's deprecated info
func (n Npm) Changelog(from, to string) *result.Changelogs {

	_, _, err := n.getVersions(context.Background())
	if err != nil {
		return nil
	}

	if len(n.data.Versions) == 0 {
		return nil
	}

	if n.data.Repository.Type == "git" && n.data.Repository.URL != "" {
		repositoryURL := n.data.Repository.URL
		if strings.HasPrefix(repositoryURL, "git+https://github.com/") {
			url := strings.TrimPrefix(repositoryURL, "git+")
			url = strings.TrimPrefix(url, "https://")
			url = strings.TrimPrefix(url, "ssh://")
			url = strings.TrimSuffix(url, ".git")
			if strings.HasPrefix(url, "git@github.com:") {
				url = strings.ReplaceAll(url, "git@github.com:", "github.com/")
			}
			if releases := vprefix.Search(from, to, func(from, to string) *result.Changelogs {
				return GetChangelogFromGitHub(url, from, to)
			}); releases != nil {
				return releases
			}
		}

		if releases := vprefix.Search(from, to, func(from, to string) *result.Changelogs {
			return giteaChangelog.SearchRepository(repositoryURL, from, to)
		}); releases != nil {
			return releases
		}
	}

	return GetChangelogFromNPMRegistry(&n.data, from, to)

}

// filterVersions filters versions based on from and to
func filterVersions(versions []string, from, to string) []string {

	if from == "" && to == "" {
		return versions
	}

	foundFrom := false
	foundTo := false

	filteredVersions := []string{}

	for _, version := range versions {
		if from != "" {
			if version == from {
				foundFrom = true
			}
		}

		if to != "" {
			if version == to {
				foundTo = true
				filteredVersions = append(filteredVersions, version)
				break
			}
		}

		if foundFrom {
			filteredVersions = append(filteredVersions, version)
		}
	}

	if len(filteredVersions) == 0 {
		return nil
	}

	if from != "" && !foundTo {
		return filteredVersions
	}

	return filteredVersions

}

// GetChangelogFromGitHub returns the changelog from a GitHub repository
func GetChangelogFromGitHub(url, from, to string) *result.Changelogs {

	parsedModule := strings.Split(url, "/")

	if len(parsedModule) < 3 {
		return nil
	}

	if parsedModule[0] != githubHost {
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      parsedModule[1],
		Repository: parsedModule[2],
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		log.Debugf("ignored error, searching releases: %s", err)
	}

	if len(releases) == 0 {
		return nil
	}

	return &releases

}

func GetChangelogFromNPMRegistry(data *Data, from, to string) *result.Changelogs {
	var changelogs result.Changelogs

	for _, i := range filterVersions(data.orderedVersions, from, to) {
		version, found := data.Versions[i]
		if !found {
			continue
		}
		changelog := result.Changelog{
			Title: version.Version,
		}

		if deprecatedMessage, ok := version.Deprecated.(string); ok {
			changelog.Body = fmt.Sprintf("Deprecated warning: %s", deprecatedMessage)
		}

		changelogs = append(changelogs, changelog)
	}

	if len(changelogs) == 0 {
		return nil
	}

	return &changelogs
}
