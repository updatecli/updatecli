package gomodule

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Target is not support for gomodule
func (g *GoModule) Target(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver, dryRun bool, releaseTarget *result.Target) error {
	return fmt.Errorf("Target not supported for the plugin GO module")
}
