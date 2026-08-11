package maven

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (m Maven) Target(_ context.Context, source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("Target not supported for the plugin Maven")
}
