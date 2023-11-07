package registry

import (
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a specific version is published
func (t *TerraformRegistry) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Debugln("scm is not supported")
	}

	versionToCheck := t.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}

	if len(versionToCheck) == 0 {
		return fmt.Errorf("%s version undefined", result.FAILURE)
	}

	versions, err := t.versions()
	if err != nil {
		return fmt.Errorf("%s retrieving terraform registry version: %w", result.FAILURE, err)
	}

	if slices.Contains(versions, versionToCheck) {
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("Terraform registry version %q available", versionToCheck)
		return nil
	}

	return fmt.Errorf("%s terraform registry version %q doesn't exist", result.FAILURE, versionToCheck)
}
