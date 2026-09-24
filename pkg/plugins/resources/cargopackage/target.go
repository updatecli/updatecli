package cargopackage

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func (cp *CargoPackage) Target(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("Target not supported for the plugin Cargo Package")
}
