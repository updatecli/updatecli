package dockerimage

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func (di *DockerImage) Target(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver, dryRun bool, resultTarget *result.Target) error {
	return fmt.Errorf("target not supported for the plugin Docker Image")
}
