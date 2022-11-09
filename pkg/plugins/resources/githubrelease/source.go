package githubrelease

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source retrieves a specific version tag from Github Releases.
func (gr *GitHubRelease) Source(workingDir string) (value string, err error) {

	versions, err := gr.ghHandler.SearchReleases(gr.typeFilter)
	if err != nil {
		logrus.Error(err)
		return value, err
	}

	if len(versions) == 0 {
		switch gr.spec.TypeFilter.IsZero() {
		case true:
			logrus.Warningf("%s No GitHub Release found, we temporary fallback to published git tags", result.ATTENTION)
			logrus.Warnln(deprecationTagSearchMessage)

			versions, err = gr.ghHandler.SearchTags()
			if err != nil {
				logrus.Errorf("%s", err)
				return "", err
			}
			if len(versions) == 0 {
				return "", fmt.Errorf("no GitHub release or git tags found, exiting")
			}
		case false:
			return "", fmt.Errorf("no GitHub release found, exiting")
		}
	}

	gr.foundVersion, err = gr.versionFilter.Search(versions)
	if err != nil {
		return "", err
	}
	value = gr.foundVersion.GetVersion()

	if len(value) == 0 {
		logrus.Infof("%s No Github Release version found matching pattern %q", result.FAILURE, gr.versionFilter.Pattern)
		return value, fmt.Errorf("no Github Release version found matching pattern %q", gr.versionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("%s Github Release version %q found matching pattern %q", result.SUCCESS, value, gr.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in Github source")
	}

	return value, nil
}
