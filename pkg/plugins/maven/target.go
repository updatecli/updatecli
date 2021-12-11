package maven

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (m Maven) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Maven")
}

func (m Maven) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Maven")
}
