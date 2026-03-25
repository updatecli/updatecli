package provider

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *TerraformProvider) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
	return fmt.Errorf("Source not supported for the plugin terraform/provider")
}
