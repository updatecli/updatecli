package gomod

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a specific stable Golang version is published
func (g *GoMod) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	var err error

	versionToCheck := g.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}

	if len(versionToCheck) == 0 {
		return errors.New("no version defined")
	}

	g.foundVersion, err = g.version(g.filename)
	if err != nil {
		if err == ErrModuleNotFound {
			return fmt.Errorf("module path %q not found", g.spec.Module)
		}

		return fmt.Errorf("looking for Golang version: %w", err)
	}

	if g.foundVersion == versionToCheck {
		switch g.kind {
		case kindGolang:
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Golang version %q found", g.foundVersion)

		case kindModule:
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Module version %q found for %q", g.foundVersion, g.spec.Module)
		}
		return nil
	}

	switch g.kind {
	case kindGolang:
		return fmt.Errorf("golang version %q found, expecting %q",
			g.foundVersion, versionToCheck)
	case kindModule:
		return fmt.Errorf("golang module version %q found for %q, expecting %q",
			g.foundVersion, g.spec.Module, versionToCheck)
	}
	return fmt.Errorf("something unexpected happened")
}
