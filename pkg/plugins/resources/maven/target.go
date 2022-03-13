package maven

import (
	"fmt"
)

func (m Maven) Target(source, workingDir string, dryRun bool) (changed bool, files []string, message string, err error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Maven")
}
