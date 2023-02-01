package githubrelease

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (gr GitHubRelease) Condition(source string) (bool, error) {
	versions, err := gr.ghHandler.SearchReleases(gr.typeFilter)

	expectedValue := source

	if gr.spec.Tag != "" {
		expectedValue = gr.spec.Tag
	}

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(versions) == 0 {
		switch gr.spec.TypeFilter.IsZero() {
		case true:
			logrus.Warningf("%s No GitHub Release found, we fallback to published git tags", result.ATTENTION)

			versions, err = gr.ghHandler.SearchTags()
			if err != nil {
				logrus.Errorf("%s", err)
				return false, err
			}
			if len(versions) == 0 {
				return false, fmt.Errorf("no GitHub release or git tags found, exiting")
			}
		case false:
			return false, fmt.Errorf("no GitHub release found, exiting")
		}
	}

	for _, version := range versions {
		if version == expectedValue {
			logrus.Infof("%s Github Release version %q found", result.SUCCESS, expectedValue)
			return true, nil
		}
	}

	return false, fmt.Errorf("%s Github Release %q not found", result.FAILURE, expectedValue)
}

func (ghr GitHubRelease) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("condition not supported for plugin GitHub Release used with scm")
}
