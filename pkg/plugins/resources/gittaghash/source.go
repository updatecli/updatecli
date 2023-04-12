package gittaghash

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag's hash based on create time
func (gth *GitTagHash) Source(workingDir string) (string, error) {

	if len(gth.spec.Path) == 0 && len(workingDir) > 0 {
		gth.spec.Path = workingDir
	}

	err := gth.Validate()
	if err != nil {
		return "", err
	}

	hashes, err := gth.nativeGitHandler.TagHashes(gth.spec.Path)

	if err != nil {
		return "", err
	}

	gth.foundVersion, err = gth.versionFilter.Search(hashes)
	if err != nil {
		return "", err
	}
	value := gth.foundVersion.GetVersion()

	if len(value) == 0 {
		logrus.Infof("%s No git tag has found matching pattern %q", result.FAILURE, gth.versionFilter.Pattern)
		return value, fmt.Errorf("no git tag has found matching pattern %q", gth.versionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("%s Git tag has %q found matching pattern %q", result.SUCCESS, value, gth.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in gitTagHash source")
	}

	return value, nil
}
