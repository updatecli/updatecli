package npm

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

// Changelog returns the link to the found npm package version's deprecated info
func (n Npm) Changelog(from, to string) *result.Changelogs {

	_, _, err := n.getVersions()
	if err != nil {
		return nil
	}

	if len(n.data.Versions) == 0 {
		return nil
	}

	if n.data.Repository.Type == "git" && n.data.Repository.URL != "" {
		if strings.HasPrefix(n.data.Repository.URL, "git+https://github.com/") {
			url := strings.TrimPrefix(n.data.Repository.URL, "git+")
			url = strings.TrimPrefix(url, "https://")
			url = strings.TrimPrefix(url, "ssh://")
			url = strings.TrimSuffix(url, ".git")
			if strings.HasPrefix(url, "git@github.com:") {
				url = strings.ReplaceAll(url, "git@github.com:", "github.com/")
			}
			if releases := GetChangelogFromGitHub(url, from, to); releases != nil {
				return releases
			}

			// Version on npm registry drops the v prefix but sometimes version on GitHub
			// contains it so we try to fetch the changelog with and without the v prefix
			vfrom := from
			vto := to
			if !strings.HasPrefix(vfrom, "v") && vfrom != "" {
				vfrom = "v" + vfrom
			}

			if !strings.HasPrefix(vto, "v") && vto != "" {
				vto = "v" + vto
			}

			if vfrom != from || vto != to {
				if releases := GetChangelogFromGitHub(url, vfrom, vto); releases != nil {
					return releases
				}
			}

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

	if parsedModule[0] != "github.com" {
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      parsedModule[1],
		Repository: parsedModule[2],
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		logrus.Debugf("ignored error, searching releases: %s", err)
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
