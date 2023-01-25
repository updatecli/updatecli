package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag based on create time
func (gt *GitTag) Source(workingDir string) (string, error) {

	if len(gt.spec.Path) == 0 && len(workingDir) > 0 {
		gt.spec.Path = workingDir
	}

	err := gt.Validate()
	if err != nil {
		return "", err
	}

	tags, err := gt.nativeGitHandler.Tags(gt.spec.Path)

	if err != nil {
		return "", err
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	if err != nil {
		return "", err
	}
	value := gt.foundVersion.GetVersion()

	if len(value) == 0 {
		logrus.Infof("%s No git tag found matching pattern %q", result.FAILURE, gt.versionFilter.Pattern)
		return value, fmt.Errorf("no git tag found matching pattern %q", gt.versionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("%s Git tag %q found matching pattern %q", result.SUCCESS, value, gt.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in gitTag source")
	}

	return value, nil
}
