package ami

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Source returns the latest AMI matching filter(s)
func (a *AMI) Source(workingDir string) (string, error) {

	svc, err := a.Init()

	if err != nil {
		return "", err
	}

	if svc == nil {
		return "", ErrWrongServiceConnection
	}

	logrus.Debugf("Looking for latest AMI ID matching:\n  ---\n  %s\n  ---\n\n",
		strings.TrimRight(
			strings.ReplaceAll(a.Spec.String(), "\n", "\n  "), "\n  "))

	result, err := a.getLatestAmiID(svc)

	if err != nil {
		return "", err
	}

	if len(result) > 0 {
		logrus.Infof("\u2714 AMI %q found\n", result)
		return result, nil
	}

	logrus.Infof("\u2717 No AMI found matching criteria in region %s\n", a.Spec.Region)

	return "", nil
}
