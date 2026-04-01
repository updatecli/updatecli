package tag

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target ensure that a specific release exist on Bitbucket Server, otherwise creates it
func (g Stash) Target(_ context.Context, source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the plugin Stash Tags")
}
