package awsami

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest AMI matching filter(s)
func (a *AMI) Source(workingDir string) (string, error) {
	logrus.Debugf("Looking for latest AMI ID matching:\n  ---\n  %s\n  ---\n\n",
		strings.TrimRight(
			strings.ReplaceAll(a.Spec.String(), "\n", "\n  "), "\n  "))

	foundAMI, err := a.getLatestAmiID()

	if err != nil {
		return "", err
	}

	if len(foundAMI) > 0 {
		logrus.Infof("%s AMI %q found\n", result.SUCCESS, foundAMI)
		return foundAMI, nil
	}

	logrus.Infof("%s No AMI found matching criteria in region %s\n", result.FAILURE, a.Spec.Region)

	return "", nil
}
