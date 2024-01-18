package registry

import (
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a specific version is published
func (t *TerraformRegistry) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Debugln("scm is not supported")
	}

	versionToCheck := t.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}

	if len(versionToCheck) == 0 {
		return false, "", fmt.Errorf("%s version undefined", result.FAILURE)
	}

	versions, err := t.versions()
	if err != nil {
		return false, "", fmt.Errorf("%s retrieving terraform registry version: %w", result.FAILURE, err)
	}

	if slices.Contains(versions, versionToCheck) {
		return true, fmt.Sprintf("Terraform registry version %q available", versionToCheck), nil
	}

	return false, fmt.Sprintf("%s terraform registry version %q doesn't exist", result.FAILURE, versionToCheck), nil
}
