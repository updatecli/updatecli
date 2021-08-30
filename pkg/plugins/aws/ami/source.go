package ami

import (
	"errors"

	"github.com/sirupsen/logrus"
)

// Source returns the latest AMI matching filter(s)
func (a *AMI) Source(workingDir string) (string, error) {

	svc, errs := a.Init()

	if len(errs) > 0 {
		for _, err := range errs {
			logrus.Printf("%s\n", err.Error())
		}
		return "", errors.New("something went wrong while retrieving ec2 AMI")
	}

	if svc == nil {
		return "", errors.New("Something went wrong while connecting AWS API")
	}

	result, err := a.getLatestAmiID(svc)

	if err != nil {
		return "", err
	}

	if len(result) > 0 {
		logrus.Infof("\u2714 AMI %q found\n", result)
		return result, nil
	}

	logrus.Infof("\u2717 No AMI found matching criteria in region %s\n", a.Region)

	logrus.Debugf("AMI Filter:\n%s\n\n", a.Filters.String())

	return "", nil
}
