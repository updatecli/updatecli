package maven

import (
	"fmt"
)

func (m Maven) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Maven")
}
