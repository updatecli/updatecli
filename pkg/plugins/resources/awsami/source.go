package awsami

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest AMI matching filter(s)
func (a *AMI) Source(workingDir string, resultSource *result.Source) error {
	logrus.Debugf("Looking for latest AMI ID matching:\n  ---\n  %s\n  ---\n\n",
		strings.TrimRight(
			strings.ReplaceAll(a.Spec.String(), "\n", "\n  "), "\n "))

	foundAMI, err := a.getLatestAmiID()

	if err != nil {
		return fmt.Errorf("get latest AMI id: %w", err)
	}

	if len(foundAMI) > 0 {

		resultSource.Result = result.SUCCESS
		resultSource.Information = foundAMI
		resultSource.Description = fmt.Sprintf("AMI %q found\n", foundAMI)
		return nil
	}

	resultSource.Result = result.FAILURE
	resultSource.Description = fmt.Sprintf("no AMI found matching criteria in region %s\n", a.Spec.Region)

	return nil
}
