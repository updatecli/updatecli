package githubrelease

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (ghr GitHubRelease) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin GitHub Release")
}

func (ghr GitHubRelease) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin GitHub Release")
}
