package jenkins

import (
	"fmt"
)

func (j Jenkins) Target(source, workingDir string, dryRun bool) (changed bool, files []string, message string, err error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Jenkins")
}
