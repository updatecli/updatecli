package gittag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/git/generic"
)

// Condition checks that a git tag exists
func (gt *GitTag) Condition(source string) (bool, error) {
	if len(gt.spec.VersionFilter.Pattern) == 0 {
		gt.spec.VersionFilter.Pattern = source
	}

	err := gt.Validate()
	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	tags, err := generic.Tags(gt.spec.Path)

	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	err = gt.spec.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}
	tag := gt.foundVersion.ParsedVersion

	if len(tag) == 0 {
		err = fmt.Errorf("no git tag matching pattern %q, found", gt.spec.VersionFilter.Pattern)
		return false, err
	}

	if strings.Compare(tag, gt.spec.VersionFilter.Pattern) == 0 {
		logrus.Printf("%s git tag %q matching\n", result.SUCCESS, gt.spec.VersionFilter.Pattern)
		return true, nil
	}

	logrus.Printf("%s git tag %q not matching %q\n",
		result.FAILURE,
		gt.spec.VersionFilter.Pattern,
		tag)

	return false, nil
}

// ConditionFromSCM test if a tag exist from a git repository specific from SCM
func (gt *GitTag) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	path := scm.GetDirectory()

	if len(gt.spec.VersionFilter.Pattern) == 0 {
		gt.spec.VersionFilter.Pattern = source
	}

	if len(gt.spec.Path) > 0 {
		logrus.Warningf("Path is defined and set to %q but is overridden by the scm definition %q",
			gt.spec.Path,
			path)
	}
	gt.spec.Path = path

	err := gt.Validate()
	if err != nil {
		return false, err
	}

	tags, err := generic.Tags(gt.spec.Path)

	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	err = gt.spec.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}
	tag := gt.foundVersion.ParsedVersion

	if tag == gt.spec.VersionFilter.Pattern {
		logrus.Printf("%s Git Tag %q matching\n", result.SUCCESS, gt.spec.VersionFilter.Pattern)
		return true, nil
	}
	logrus.Printf("%s Git Tag %q not matching %q\n",
		result.FAILURE,
		gt.spec.VersionFilter.Pattern,
		tag)
	return false, nil
}
