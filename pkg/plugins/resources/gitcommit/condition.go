package gitcommit

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition is not supported for the Git Commit resource.
func (gc *GitCommit) Condition(_ context.Context, source string, scm scm.ScmHandler) (bool, string, error) {
	return false, "", fmt.Errorf("condition not supported for the Git Commit resource")
}
