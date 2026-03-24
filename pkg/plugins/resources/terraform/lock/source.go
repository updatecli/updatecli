package lock

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *TerraformLock) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {
	return fmt.Errorf("Source not supported for the plugin terraform/lock")
}
