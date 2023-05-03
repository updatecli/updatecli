package githubrelease

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (gr GitHubRelease) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("condition not supported for plugin GitHub Release used with scm")
	}

	expectedValue := source
	if gr.spec.Tag != "" {
		expectedValue = gr.spec.Tag
	}

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
				return fmt.Errorf("looking for GitHub release tag: %w", err)
			}
			if len(versions) == 0 {
				return fmt.Errorf("no GitHub release or git tags found")
			}
		case false:
			return fmt.Errorf("no GitHub release found")
		}
	}

	for _, version := range versions {
		if version == expectedValue {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("GitHub release %q found", expectedValue)
			return nil
		}
	}

	return fmt.Errorf("GitHub release %q not found", expectedValue)
}
