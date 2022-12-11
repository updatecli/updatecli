package npm

import (
	"fmt"
)

func (n Npm) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Npm")
}
