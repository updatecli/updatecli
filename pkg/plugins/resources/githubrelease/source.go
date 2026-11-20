package githubrelease

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Source retrieves a specific version tag name, tag hash, or release title from GitHub Releases.
func (gr *GitHubRelease) Source(ctx context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {

	releaseRefs, err := gr.ghHandler.SearchReleases(ctx, gr.typeFilter, 0)
	if err != nil {
		return err
	}

	// published holds every release, versions the ones within the age window
	var published, versions []string
	for _, release := range releaseRefs {
		published = append(published, release.TagName)
		date := release.PublicationDate()
		if !gr.spec.Age.Matches(date) {
			log.Debugf("ignoring release %q, published on %s, as outside of the age window", release.TagName, date)
			continue
		}
		versions = append(versions, release.TagName)
	}
	/*
		The git tag fallback only applies to a repository without any release. When the age
		filter discarded every release, it's deliberately not attempted: it only reports tag
		names, so it couldn't honor the age window.
	*/
	if len(releaseRefs) == 0 {
		switch gr.spec.TypeFilter.IsZero() {
		case true:
			log.Warningf("%s No GitHub Release found, we fallback to published git tags", result.ATTENTION)
			if !gr.spec.Age.IsZero() {
				log.Warningf("%s The age filter cannot be honored by the git tag fallback, as git tags are reported without their date", result.ATTENTION)
			}

			versions, err = gr.ghHandler.SearchTags(ctx, 0)
			if err != nil {
				return fmt.Errorf("searching git tag: %w", err)
			}
			if len(versions) == 0 {
				return fmt.Errorf("no GitHub release or git tags found, exiting")
			}
			// Git tags are reported without their date, so the age filter doesn't apply to them
			published = versions
		case false:
			return fmt.Errorf("no GitHub release found, exiting")
		}
	}

	gr.foundVersion, err = age.Search(&gr.versionFilter, published, versions)
	if err != nil {
		// Every release matching the version filter may still be cooling down
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			log.Debugf("%s", err)
			resultSource.Result = result.SKIPPED
			resultSource.Description = "no GitHub release matching the version filter matches the age filter yet"
			return nil
		}
		return fmt.Errorf("filtering github release version: %w", err)
	}

	value := gr.foundVersion.GetVersion()

	if gr.spec.Key == KeyTagHash {
		for _, release := range releaseRefs {
			if release.TagName == value {
				value = release.TagCommit.Oid
			}
		}
	}

	if gr.spec.Key == KeyTitle {
		for _, release := range releaseRefs {
			value = release.Name
		}
	}

	if len(value) == 0 {
		return fmt.Errorf("no GitHub Release version found matching pattern %q of kind %q",
			gr.versionFilter.Pattern,
			gr.versionFilter.Kind,
		)

	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("GitHub release version %q found matching pattern %q of kind %q",
		value,
		gr.versionFilter.Pattern,
		gr.versionFilter.Kind)

	return nil
}
