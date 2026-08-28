package hcl

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

func (h *Hcl) Source(_ context.Context, resolver utils.Resolver, resultSource *result.Source) error {
	if len(h.files) > 1 {
		return fmt.Errorf("%s HCL source only supports one file", result.FAILURE)
	}

	if err := h.UpdateAbsoluteFilePath(resolver); err != nil {
		return err
	}

	if err := h.Read(); err != nil {
		return fmt.Errorf("reading hcl file: %w", err)
	}

	// Always one
	var filePath string
	for f := range h.files {
		filePath = f
	}

	resourceFile := h.files[filePath]
	sourceOutput, err := h.Query(resourceFile)
	if err != nil {
		return err
	}

	resultSource.Information = sourceOutput
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for path %q'",
		sourceOutput,
		resourceFile.originalFilePath,
		h.spec.Path)

	return nil
}
