package gomodule

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a go module with a specific version is published
func (g *GoModule) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	versionToCheck := g.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return errors.New("no version defined")
	}

	_, versions, err := g.versions()
	if err != nil {
		return fmt.Errorf("searching version: %w", err)
	}

	for _, v := range versions {
		if v == versionToCheck {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("version %q available", versionToCheck)
			return nil
		}
	}

	return fmt.Errorf("version %q doesn't exist", versionToCheck)
}
