package branch

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Target ensure that a specific release exist on GitLab, otherwise creates it
func (g Gitlab) Target(_ context.Context, source string, scm scm.ScmHandler, resolver utils.Resolver, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the plugin GitLab branch")
}
