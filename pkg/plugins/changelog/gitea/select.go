package gitea

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/drone/go-scm/scm"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// selectReleases orders and selects releases for the changelog.
func selectReleases(releases []*scm.Release, filter version.Filter, from, to string) []*scm.Release {
	useSemver := false
	parseVersion := semver.NewVersion
	switch filter.Kind {
	case version.SEMVERVERSIONKIND, version.REGEXSEMVERVERSIONKIND:
		useSemver = true
		if filter.Strict {
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
		log.Debugf("Gitea changelog for version kind %q only includes release %q", filter.Kind, to)
		return filterReleases(releases, to, to)
	}
	if !useSemver {
		return filterReleases(releases, from, to)
	}

	var versionRegex, replaceRegex *regexp.Regexp
	if filter.Kind == version.REGEXSEMVERVERSIONKIND {
		var err error
		versionRegex, err = regexp.Compile(filter.Regex)
		if err != nil {
			log.Debugf("failed to compile Gitea changelog version regex: %s", err)
			return nil
		}
		if filter.ReplaceAll.Pattern != "" {
			replaceRegex, err = regexp.Compile(filter.ReplaceAll.Pattern)
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
			value = replaceRegex.ReplaceAllString(value, filter.ReplaceAll.Replacement)
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
		// Without either version there is no range, and returning every release would flood the changelog.
		return nil
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
		// Without a current version, listing every older release could flood the changelog.
		return releases[toIndex : toIndex+1]
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
