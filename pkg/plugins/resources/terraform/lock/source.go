package lock

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

func (t *TerraformLock) Source(_ context.Context, resolver utils.Resolver, resultSource *result.Source) error {
	return fmt.Errorf("Source not supported for the plugin terraform/lock")
}
