package pypi

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

// Changelog returns release notes for the package between the from and to versions.
func (p *Pypi) Changelog(from, to string) *result.Changelogs {
	// Only the package metadata is needed, which the source already fetched in the same run.
	// Searching versions again would repeat the age filter, whose outcome must not prevent
	// the changelog.
	if p.data.Info.Name == "" && len(p.data.Releases) == 0 {
		data, err := p.getPackageData(context.Background())
		if err != nil {
			return nil
		}
		p.data = data
	}

	// Prefer forge releases when a Source URL is provided in project_urls.
	if sourceURL, ok := p.data.Info.ProjectURLs["Source"]; ok {
		for _, search := range []func(rawURL, from, to string) *result.Changelogs{
			changelogFromGitHub,
			giteaChangelog.SearchRepository,
		} {
			// PyPI versions rarely have the "v" prefix but release tags often do.
			if releases := vprefix.Search(from, to, func(from, to string) *result.Changelogs {
				return search(sourceURL, from, to)
			}); releases != nil {
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
		log.Debugf("searching GitHub releases for %s/%s: %s", parts[1], parts[2], err)
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
