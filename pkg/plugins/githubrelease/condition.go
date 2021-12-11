package githubrelease

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (ghr GitHubRelease) Condition(source string) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}

func (ghr GitHubRelease) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}
