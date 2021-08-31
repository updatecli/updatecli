package ami

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

// Source returns the latest AMI matching filter(s)
func (a *AMI) Source(workingDir string) (string, error) {

	svc, errs := a.Init()

	if len(errs) > 0 {
		return "", errors.New("something went wrong while retrieving ec2 AMI")
	}

	if svc == nil {
		return "", errors.New("Something went wrong while connecting AWS API")
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
