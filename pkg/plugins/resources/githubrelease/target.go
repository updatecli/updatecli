package githubrelease

import (
	"fmt"
)

func (ghr GitHubRelease) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin GitHub Release")
}
