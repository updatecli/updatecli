package pypi

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target is not supported for the pypi plugin.
func (p *Pypi) Target(source string, scmHandler scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the pypi plugin")
}
