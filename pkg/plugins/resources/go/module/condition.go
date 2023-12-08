package gomodule

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a go module with a specific version is published
func (g *GoModule) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	versionToCheck := g.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, "", fmt.Errorf("no version defined")
	}

	_, versions, err := g.versions()
	if err != nil {
		return false, "", fmt.Errorf("searching version: %w", err)
	}

	for _, v := range versions {
		if v == versionToCheck {
			return true, fmt.Sprintf("version %q available", versionToCheck), nil
		}
	}

	return false, fmt.Sprintf("version %q doesn't exist", versionToCheck), nil
}
