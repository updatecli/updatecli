package provider

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func (t *TerraformProvider) Source(_ context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	return fmt.Errorf("Source not supported for the plugin terraform/provider")
}
