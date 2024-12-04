package gomodule

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target is not support for gomodule
func (g *GoModule) Target(source result.SourceInformation, scm scm.ScmHandler, dryRun bool, releaseTarget *result.Target) error {
	return fmt.Errorf("Target not supported for the plugin GO module")
}
