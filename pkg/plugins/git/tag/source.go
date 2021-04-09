package tag

import (
	"fmt"

	"github.com/olblak/updateCli/pkg/plugins/git/generic"
	"github.com/sirupsen/logrus"
)

// Source return the latest git tag based on create time
func (t *Tag) Source(workingDir string) (string, error) {

	if len(t.Path) == 0 && len(workingDir) > 0 {
		t.Path = workingDir
	}

	err := t.Validate()
	if err != nil {
		logrus.Errorln(err)
		return "", err
	}

	tags, err := generic.Tags(workingDir)

	if err != nil {
		logrus.Errorln(err)
		return "", err
	}

	value, err := t.VersionFilter.Search(tags)
	if err != nil {
		return "", err
	}

	if len(value) == 0 {
		logrus.Infof("\u2717 No Git Tag found matching pattern %q", t.VersionFilter.Pattern)
		return value, fmt.Errorf("no Git tag found matching pattern %q", t.VersionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("\u2714 Git Tag %q found, matching pattern %q", value, t.VersionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in gitTag source")
	}

	return value, nil
}
