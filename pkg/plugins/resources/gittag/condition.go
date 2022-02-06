package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Condition checks that a git tag exists
func (gt *GitTag) Condition(source string) (bool, error) {
	return gt.condition(source)
}

// ConditionFromSCM test if a tag exists from a git repository specific from SCM
func (gt *GitTag) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	path := scm.GetDirectory()

	if len(gt.spec.Path) > 0 {
		logrus.Warningf("Path is defined and set to %q but is overridden by the scm definition %q",
			gt.spec.Path,
			path)
	}
	gt.spec.Path = path

	return gt.condition(source)
}

func (gt *GitTag) condition(source string) (bool, error) {
	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if source != "" {
		logrus.Infof("Source input value detected: using it as spec.versionfilter.pattern")
		gt.spec.VersionFilter.Pattern = source
	}

	err := gt.Validate()
	if err != nil {
		return false, err
	}

	tags, err := gitgeneric.Tags(gt.spec.Path)
	if err != nil {
		return false, err
	}

	gt.foundVersion, err = gt.spec.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}
	tag := gt.foundVersion.ParsedVersion

	if len(tag) == 0 {
		err = fmt.Errorf("No git tag matching pattern %q, found", gt.spec.VersionFilter.Pattern)
		return false, err
	}

	if tag == gt.spec.VersionFilter.Pattern {
		logrus.Printf("%s Git tag %q matching\n", result.SUCCESS, gt.spec.VersionFilter.Pattern)
		return true, nil
	}

	logrus.Printf("%s Git Tag %q not matching %q\n",
		result.FAILURE,
		gt.spec.VersionFilter.Pattern,
		tag)

	return false, nil
}
