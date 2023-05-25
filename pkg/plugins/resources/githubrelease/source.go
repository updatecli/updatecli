package githubrelease

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source retrieves a specific version tag from GitHub Releases.
func (gr *GitHubRelease) Source(workingDir string, resultSource *result.Source) error {

	versions, err := gr.ghHandler.SearchReleases(gr.typeFilter)
	if err != nil {
		return fmt.Errorf("searching GitHub release: %w", err)
	}

	if len(versions) == 0 {
		switch gr.spec.TypeFilter.IsZero() {
		case true:
			logrus.Warningf("%s No GitHub Release found, we fallback to published git tags", result.ATTENTION)

			versions, err = gr.ghHandler.SearchTags()
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
