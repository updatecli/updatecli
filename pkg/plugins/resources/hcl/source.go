package hcl

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (h *Hcl) Source(workingDir string, resultSource *result.Source) error {
	if len(h.files) > 1 {
		return fmt.Errorf("%s HCL source only supports one file", result.FAILURE)
	}

	h.UpdateAbsoluteFilePath(workingDir)

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
