package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a git tag exists
func (gt *GitTag) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		path := scm.GetDirectory()

		if len(gt.spec.Path) > 0 {
			logrus.Debugf("path is defined and set to %q but is overridden by the scm definition %q",
				gt.spec.Path,
				path)
		}
		gt.spec.Path = path
	}

	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if source != "" {
		logrus.Infof("Source input value detected: using it as spec.versionfilter.pattern")
		gt.versionFilter.Pattern = source
	}

	err := gt.Validate()
	if err != nil {
		return err
	}

	tags, err := gt.nativeGitHandler.Tags(gt.spec.Path)
	if err != nil {
		return err
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	if err != nil {
		return err
	}
	tag := gt.foundVersion.GetVersion()

	if len(tag) == 0 {
		return fmt.Errorf("no git tag matching pattern %q, found", gt.versionFilter.Pattern)
	}

	if tag == gt.versionFilter.Pattern {
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("git tag %q matching\n", gt.versionFilter.Pattern)
		return nil
	}

	return fmt.Errorf("git tag %q not matching %q", gt.versionFilter.Pattern, tag)
}
