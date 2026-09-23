package pypi

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Target is not supported for the pypi plugin.
func (p *Pypi) Target(_ context.Context, source string, scmHandler scm.ScmHandler, pathResolver pathresolver.Resolver, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the pypi plugin")
}
