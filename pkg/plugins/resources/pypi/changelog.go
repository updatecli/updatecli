package pypi

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

// Changelog returns release notes for the package between the from and to versions.
func (p *Pypi) Changelog(from, to string) *result.Changelogs {
	_, _, err := p.getVersions(context.Background())
	if err != nil {
		return nil
	}

	// Prefer GitHub releases when a Source URL is provided in project_urls.
	if githubURL, ok := p.data.Info.ProjectURLs["Source"]; ok {
		if releases := changelogFromGitHub(githubURL, from, to); releases != nil {
			return releases
		}

		// PyPI versions rarely have the "v" prefix but GitHub tags often do.
		vfrom, vto := withVPrefix(from), withVPrefix(to)
		if vfrom != from || vto != to {
			if releases := changelogFromGitHub(githubURL, vfrom, vto); releases != nil {
				return releases
			}
		}
	}

	return changelogFromPyPI(p.spec.Name, p.spec.URL, from, to)
}

// changelogFromGitHub extracts owner/repo from a GitHub URL and queries the GitHub release API.
func changelogFromGitHub(rawURL, from, to string) *result.Changelogs {
	// Normalise: strip scheme and trailing .git
	url := rawURL
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	// Expect: github.com / owner / repo
	if len(parts) < 3 || parts[0] != "github.com" {
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      parts[1],
		Repository: parts[2],
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		logrus.Debugf("searching GitHub releases for %s/%s: %s", parts[1], parts[2], err)
	}

	if len(releases) == 0 {
		return nil
	}

	return &releases
}

// changelogFromPyPI builds fallback changelog entries pointing to PyPI release pages.
func changelogFromPyPI(packageName, baseURL, from, to string) *result.Changelogs {
	var changelogs result.Changelogs

	versions := []string{from}
	if to != from && to != "" {
		versions = append(versions, to)
	}

	for _, ver := range versions {
		if ver == "" {
			continue
		}
		changelogs = append(changelogs, result.Changelog{
			Title: ver,
			URL:   fmt.Sprintf("%spypi/%s/%s/", baseURL, packageName, ver),
		})
	}

	if len(changelogs) == 0 {
		return nil
	}

	return &changelogs
}

// withVPrefix prepends "v" if the version string does not already have it.
func withVPrefix(ver string) string {
	if ver == "" || strings.HasPrefix(ver, "v") {
		return ver
	}
	return "v" + ver
}
