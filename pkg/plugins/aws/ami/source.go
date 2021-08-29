package ami

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Source returns the latest AMI matching the filter
func (a *AMI) Source(workingDir string) (string, error) {

	svc, errs := a.Init()

	for _, err := range errs {
		logrus.Error(err)
	}
	if len(errs) > 0 {
		return "", errors.New("Too many errors")
	}

	if svc == nil {
		return "", errors.New("Something went wrong in aws connection")
	}

	result, err := a.GetLatestAmiID(svc)

	if err != nil {
		return "", err
	}

	if len(result) > 0 {
		logrus.Infof("\u2714 AMI %q found\n", result)
		return result, nil
	}

	fmt.Printf("\u2717 No AMI found matching criteria in region %s\n", a.Region)

	logrus.Debugf("AMI Filter:\n%s\n\n", a.Filters.String())

	return "", nil
}
