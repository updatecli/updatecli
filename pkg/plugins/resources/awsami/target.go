package awsami

import (
	"fmt"
)

func (a *AMI) Target(source, workingDir string, dryRun bool) (changed bool, files []string, message string, err error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin AWS/AMI")
}
