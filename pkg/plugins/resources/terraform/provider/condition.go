package provider

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *TerraformProvider) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if len(t.files) > 1 {
		return fmt.Errorf("%s terraform/lock condition only supports one file", result.FAILURE)
	}

	if scm != nil {
		t.UpdateAbsoluteFilePath(scm.GetDirectory())
	}

	if err := t.Read(); err != nil {
		return err
	}

	// Always one
	var filePath string
	for f := range t.files {
		filePath = f
	}

	resourceFile := t.files[filePath]
	conditionOutputVersion, err := t.Query(resourceFile)
	if err != nil {
		return err
	}

	value := source
	if t.spec.Value != "" {
		value = t.spec.Value
	}

	if value == conditionOutputVersion {
		resultCondition.Description = fmt.Sprintf("Path %q, from file %q, is correctly set to %q",
			t.spec.Provider,
			resourceFile.originalFilePath,
			value)

		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS

		return nil
	}

	resultCondition.Description = fmt.Sprintf("Path %q, from file %q, is incorrectly set to %q and should be %q",
		t.spec.Provider,
		resourceFile.originalFilePath,
		conditionOutputVersion,
		value,
	)
	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE

	return nil
}
