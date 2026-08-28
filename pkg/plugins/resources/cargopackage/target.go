package cargopackage

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

func (cp *CargoPackage) Target(_ context.Context, source string, scm scm.ScmHandler, resolver utils.Resolver, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("Target not supported for the plugin Cargo Package")
}
