package jenkins

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j Jenkins) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("Target not supported for the plugin Jenkins")
}
