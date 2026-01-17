package bazelregistry

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target is not supported for the bazelregistry resource
func (b *Bazelregistry) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the plugin bazelregistry")
}
