package provider

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *TerraformProvider) Source(workingDir string, resultSource *result.Source) error {
	return fmt.Errorf("Source not supported for the plugin terraform/provider")
}
