package githubrelease

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// Source retrieves a specific version tag name, tag hash, or release title from GitHub Releases.
func (gr *GitHubRelease) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {

	releaseRefs, err := gr.ghHandler.SearchReleases(ctx, gr.typeFilter, 0)
	if err != nil {
		return err
	}

	var versions []string
	for _, release := range releaseRefs {
		date := release.PublicationDate()
		if !gr.spec.Age.Matches(date) {
			logrus.Debugf("ignoring release %q, published on %s, as outside of the age window", release.TagName, date)
			continue
		}
		versions = append(versions, release.TagName)
	}
	/*
		The repository does publish releases but the age filter discarded every one of
		them, which is an expected state of the filter rather than a failure, so the
		source is skipped instead. The git tag fallback is deliberately not attempted
		here: it only reports tag names, so it couldn't honor the age window.
	*/
	if len(releaseRefs) > 0 && len(versions) == 0 {
		logrus.Debugf("%s", age.ErrNoVersionMatchingAge)
		resultSource.Result = result.SKIPPED
		resultSource.Description = "no GitHub release matches the age filter yet"
		return nil
	}

	if len(versions) == 0 {
		switch gr.spec.TypeFilter.IsZero() {
		case true:
			logrus.Warningf("%s No GitHub Release found, we fallback to published git tags", result.ATTENTION)
			if !gr.spec.Age.IsZero() {
				logrus.Warningf("%s The age filter cannot be honored by the git tag fallback, as git tags are reported without their date", result.ATTENTION)
			}

			versions, err = gr.ghHandler.SearchTags(ctx, 0)
			if err != nil {
				return fmt.Errorf("searching git tag: %w", err)
			}
			if len(versions) == 0 {
				return fmt.Errorf("no GitHub release or git tags found, exiting")
			}
		case false:
			return fmt.Errorf("no GitHub release found, exiting")
		}
	}

	gr.foundVersion, err = gr.versionFilter.Search(versions)
	if err != nil {
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
