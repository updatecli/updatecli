package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/git/generic"
)

// Source return the latest git tag based on create time
func (gt *GitTag) Source(workingDir string) (string, error) {

	if len(gt.spec.Path) == 0 && len(workingDir) > 0 {
		gt.spec.Path = workingDir
	}

	err := gt.Validate()
	if err != nil {
		return "", err
	}

	tags, err := generic.Tags(workingDir)

	if err != nil {
		return "", err
	}

	gt.foundVersion, err = gt.spec.VersionFilter.Search(tags)
	if err != nil {
		return "", err
	}
	value := gt.foundVersion.ParsedVersion

	if len(value) == 0 {
		logrus.Infof("%s No Git Tag found matching pattern %q", result.FAILURE, gt.spec.VersionFilter.Pattern)
		return value, fmt.Errorf("no Git tag found matching pattern %q", gt.spec.VersionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("%s Git Tag %q found, matching pattern %q", result.SUCCESS, value, gt.spec.VersionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in gitTag source")
	}

	return value, nil
}
