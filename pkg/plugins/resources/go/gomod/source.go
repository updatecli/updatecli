package gomod

import (
	"errors"
	"fmt"
	"os"

	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Source returns the latest go module version
func (g *GoMod) Source(workingDir string, resultSource *result.Source) error {
	var err error

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return errors.New("fail getting current working directory")
	}

	filename := g.filename
	// To merge File path with current working dire, unless file is an http url
	if workingDir != currentWorkingDirectory {
		filename = utils.JoinFilePathWithWorkingDirectoryPath(filename, workingDir)
	}

	g.foundVersion, err = g.version(filename)
	if err != nil {
		return fmt.Errorf("searching version: %w", err)
	}

	if g.foundVersion == "" {
		return fmt.Errorf("no version found for module path %q", g.spec.Module)
	}

	resultSource.Information = g.foundVersion
	resultSource.Result = result.SUCCESS

	switch g.kind {
	case kindGolang:
		resultSource.Description = fmt.Sprintf("Golang Version %s found", g.foundVersion)
	case kindModule:
		resultSource.Description = fmt.Sprintf("version %s found for GO module %q",
			g.foundVersion,
			g.spec.Module,
		)
	}

	return nil
}
