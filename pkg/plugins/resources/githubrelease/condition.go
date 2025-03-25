package githubrelease

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (gr GitHubRelease) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if scm != nil {
		logrus.Warningf("condition not supported for plugin GitHub Release used with scm")
	}

	expectedValue := source
	if gr.spec.Tag != "" {
		expectedValue = gr.spec.Tag
	}

	var versions []string
	if gr.spec.Key == KeyHash {
		versions, err = gr.ghHandler.SearchReleasesByTagHash(gr.typeFilter)
	} else if gr.spec.Key == KeyTitle {
		versions, err = gr.ghHandler.SearchReleasesByTitle(gr.typeFilter)
	} else {
		versions, err = gr.ghHandler.SearchReleasesByTagName(gr.typeFilter)
	}
	if err != nil {
		return false, "", fmt.Errorf("searching GitHub release: %w", err)
	}

	if len(versions) == 0 {
		switch gr.spec.TypeFilter.IsZero() {
		case true:
			logrus.Warningf("%s No GitHub Release found, we fallback to published git tags", result.ATTENTION)

			versions, err = gr.ghHandler.SearchTags()
			if err != nil {
				return false, "", fmt.Errorf("looking for GitHub release tag: %w", err)
			}
			if len(versions) == 0 {
				return false, "", fmt.Errorf("no GitHub release or git tags found")
			}
		case false:
			return false, "", fmt.Errorf("no GitHub release found")
		}
	}

	for _, version := range versions {
		if version == expectedValue {
			return true, fmt.Sprintf("GitHub release %q found", expectedValue), nil
		}
	}

	return false, fmt.Sprintf("GitHub release %q not found", expectedValue), nil
}
