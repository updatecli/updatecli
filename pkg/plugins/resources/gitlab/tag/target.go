package tag

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Target ensure that a specific release exist on GitLab, otherwise creates it
func (g Gitlab) Target(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver, dryRun bool, releaseTarget *result.Target) error {
	return fmt.Errorf("target not supported for the plugin GitLab Tags")
}
