package hcl

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (h *Hcl) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if len(h.files) > 1 {
		return fmt.Errorf("%s HCL condition only supports one file", result.FAILURE)
	}

	if scm != nil {
		h.UpdateAbsoluteFilePath(scm.GetDirectory())
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
	conditionOutput, err := h.Query(resourceFile)
	if err != nil {
		return err
	}

	value := source
	if h.spec.Value != "" {
		value = h.spec.Value
	}

	if value == conditionOutput {
		resultCondition.Description = fmt.Sprintf("Path %q, from file %q, is correctly set to %q",
			h.spec.Key,
			resourceFile.originalFilePath,
			value)

		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS

		return nil
	}

	resultCondition.Description = fmt.Sprintf("Path %q, from file %q, is incorrectly set to %q and should be %q",
		h.spec.Key,
		resourceFile.originalFilePath,
		conditionOutput,
		value,
	)
	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE

	return nil
}
