package temurin

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target is not implemented. If you ever feel the need, you can still open a GitHub issue with a valid usecase.
func (t *Temurin) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("Target not supported for resource of kind 'temurin'")
}
