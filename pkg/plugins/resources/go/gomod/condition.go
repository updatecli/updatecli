package gomod

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Condition checks if a specific stable Golang version is published
func (g *GoMod) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	versionToCheck := g.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}

	filename := g.filename
	if scm != nil {
		filename = utils.JoinFilePathWithWorkingDirectoryPath(filename, scm.GetDirectory())
	}

	g.foundVersion, err = g.version(filename)
	if err != nil {
		if err == ErrModuleNotFound {
			return false, "", fmt.Errorf("module path %q not found", g.spec.Module)
		}

		return false, "", fmt.Errorf("looking for Golang version: %w", err)
	}

	if versionToCheck == "" && g.foundVersion != "" {
		switch g.kind {
		case kindGolang:
			return true, fmt.Sprintf("Golang version %q found", g.foundVersion), nil

		case kindModule:
			return true, fmt.Sprintf("Module %q found", g.spec.Module), nil
		}
	} else if g.foundVersion == versionToCheck {
		switch g.kind {
		case kindGolang:
			return true, fmt.Sprintf("Golang version %q found", g.foundVersion), nil

		case kindModule:
			return true, fmt.Sprintf("Module version %q found for %q", g.foundVersion, g.spec.Module), nil
		}
	}

	switch g.kind {
	case kindGolang:
		return false, fmt.Sprintf("golang version %q found, expecting %q",
			g.foundVersion, versionToCheck), nil
	case kindModule:
		return false, fmt.Sprintf("golang module version %q found for %q, expecting %q",
			g.foundVersion, g.spec.Module, versionToCheck), nil
	}
	return false, "", fmt.Errorf("something unexpected happened")
}
