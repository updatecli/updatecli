package gomod

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Source returns the latest go module version
func (g *GoMod) Source(workingDir string, resultSource *result.Source) error {
	var err error

	g.foundVersion, err = g.version(utils.JoinFilePathWithWorkingDirectoryPath(g.filename, workingDir))
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
