package jenkins

import (
	"fmt"
)

func (j Jenkins) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Jenkins")
}
