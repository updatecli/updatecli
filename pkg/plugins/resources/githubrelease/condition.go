package githubrelease

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (gr GitHubRelease) Condition(source string) (bool, error) {
	versions, err := gr.ghHandler.SearchReleases(gr.spec.Type)

	expectedValue := source

	if len(gr.spec.Tag) == 0 {
		expectedValue = gr.spec.Tag
	}

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(versions) == 0 {
		switch gr.spec.DisableTagSearch {
		case true:
			return false, fmt.Errorf("no GitHub release found, exiting")
		case false:
			logrus.Infof("%s No GitHub Release found. As fallback Looking at published git tags", result.ATTENTION)
			versions, err = gr.ghHandler.SearchTags()
			if err != nil {
				logrus.Errorf("%s", err)
				return false, err
			}
			if len(versions) == 0 {
				return false, fmt.Errorf("no GitHub release or git tags found, exiting")
			}
		}
	}

	gr.foundVersion, err = gr.versionFilter.Search(versions)
	if err != nil {
		return false, err
	}

	value := gr.foundVersion.GetVersion()

	if len(value) == 0 {
		logrus.Infof("%s No Github Release version found matching pattern %q", result.FAILURE, gr.versionFilter.Pattern)
		return false, fmt.Errorf("no Github Release version found matching pattern %q", gr.versionFilter.Pattern)
	}

	if value == expectedValue {
		logrus.Infof("%s Github Release version %q found matching pattern %q", result.SUCCESS, value, gr.versionFilter.Pattern)
		return true, nil
	}

	return false, fmt.Errorf("something unexpected happened in Github source")
}

func (ghr GitHubRelease) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("condition not supported for plugin GitHub Release used with scm")
}
