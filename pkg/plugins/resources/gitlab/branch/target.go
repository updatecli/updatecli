package branch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target ensure that a specific release exist on GitLab, otherwise creates it
func (g Gitlab) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the plugin GitLab branch")
}
