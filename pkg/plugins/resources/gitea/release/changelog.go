package release

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/drone/go-scm/scm"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Changelog returns release notes between two versions, including both endpoints.
// For version kinds without range support, it returns only the new release.
func (g *Gitea) Changelog(from, to string) *result.Changelogs {
	releases, err := g.listReleases(context.Background())
	if err != nil {
		log.Debugf("failed to retrieve Gitea changelog: %s", err)
		return nil
	}

	releases = slices.DeleteFunc(releases, func(r *scm.Release) bool { return r.Draft })
	releases = g.selectReleases(releases, from, to)
	if len(releases) == 0 {
		return nil
	}

	baseURL, err := url.Parse(g.spec.URL)
	if err != nil {
		log.Debugln("failed to parse Gitea changelog URL")
		return nil
	}
	baseURL.User = nil
	releaseURLPrefix := fmt.Sprintf("%s/%s/%s/releases/tag/",
		strings.TrimRight(baseURL.String(), "/"),
		url.PathEscape(g.spec.Owner),
		url.PathEscape(g.spec.Repository),
	)
	var changelogs result.Changelogs
	for _, release := range releases {
		link := release.Link
		if link == "" {
			// go-scm may return an empty link even when html_url is present.
			link = releaseURLPrefix + url.PathEscape(release.Tag)
		}
		changelogs = append(changelogs, result.Changelog{
			Title:       release.Tag,
			Body:        release.Description,
			PublishedAt: releaseDate(release).String(),
			URL:         link,
		})
	}

	return &changelogs
}

// selectReleases orders and selects releases for the changelog.
func (g *Gitea) selectReleases(releases []*scm.Release, from, to string) []*scm.Release {
	useSemver := false
	parseVersion := semver.NewVersion
	switch g.versionFilter.Kind {
	case version.SEMVERVERSIONKIND, version.REGEXSEMVERVERSIONKIND:
		useSemver = true
		if g.versionFilter.Strict {
			parseVersion = semver.StrictNewVersion
		}
	case "", version.LATESTVERSIONKIND:
		// Strict parsing avoids mistaking date tags such as 2026-01-02 for semver.
		_, fromErr := semver.StrictNewVersion(strings.TrimPrefix(from, "v"))
		_, toErr := semver.StrictNewVersion(strings.TrimPrefix(to, "v"))
		useSemver = fromErr == nil && toErr == nil
	case version.REGEXVERSIONKIND:
		// Regex filters keep the API order used by source selection.
	default:
		if to == "" {
			return nil
		}
		log.Debugf("Gitea changelog for version kind %q only includes release %q", g.versionFilter.Kind, to)
		return filterReleases(releases, to, to)
	}
	if !useSemver {
		return filterReleases(releases, from, to)
	}

	var versionRegex, replaceRegex *regexp.Regexp
	if g.versionFilter.Kind == version.REGEXSEMVERVERSIONKIND {
		var err error
		versionRegex, err = regexp.Compile(g.versionFilter.Regex)
		if err != nil {
			log.Debugf("failed to compile Gitea changelog version regex: %s", err)
			return nil
		}
		if g.versionFilter.ReplaceAll.Pattern != "" {
			replaceRegex, err = regexp.Compile(g.versionFilter.ReplaceAll.Pattern)
			if err != nil {
				log.Debugf("failed to compile Gitea changelog replacement regex: %s", err)
				return nil
			}
		}
	}

	type semverRelease struct {
		release *scm.Release
		version *semver.Version
	}
	var semverReleases []semverRelease
	for _, release := range releases {
		value := release.Tag
		if replaceRegex != nil {
			value = replaceRegex.ReplaceAllString(value, g.versionFilter.ReplaceAll.Replacement)
		}
		if versionRegex != nil {
			matches := versionRegex.FindStringSubmatch(value)
			if len(matches) < 2 {
				continue
			}
			value = matches[1]
		}
		v, err := parseVersion(value)
		if err != nil {
			continue
		}
		semverReleases = append(semverReleases, semverRelease{release: release, version: v})
	}
	sort.SliceStable(semverReleases, func(i, j int) bool {
		return semverReleases[i].version.GreaterThan(semverReleases[j].version)
	})
	results := make([]*scm.Release, len(semverReleases))
	for i, release := range semverReleases {
		results[i] = release.release
	}
	return filterReleases(results, from, to)
}

// filterReleases selects an inclusive range from releases ordered newest first.
func filterReleases(releases []*scm.Release, from, to string) []*scm.Release {
	if from == "" && to == "" {
		return releases
	}

	fromIndex, toIndex := -1, -1
	for i, release := range releases {
		if release.Tag == from {
			fromIndex = i
		}
		if release.Tag == to {
			toIndex = i
		}
	}

	if to == "" {
		if fromIndex < 0 {
			return nil
		}
		return releases[fromIndex : fromIndex+1]
	}
	if toIndex < 0 {
		return nil
	}
	if from == "" {
		return releases[toIndex:]
	}
	if fromIndex < 0 {
		log.Debugf("Gitea release version %q not found, returning only %q", from, to)
		return releases[toIndex : toIndex+1]
	}
	if fromIndex < toIndex {
		// Match the GitHub changelog fallback for downgrades.
		return releases[fromIndex : fromIndex+1]
	}
	return releases[toIndex : fromIndex+1]
}
