package gomodule

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (g *GoModule) Source(workingDir string, resultSource *result.Source) error {
	version, _, err := g.versions()
	if err != nil {
		return fmt.Errorf("searching go module version: %w", err)
	}

	g.Version.OriginalVersion = version
	g.Version.ParsedVersion = version

	if version == "" {
		return fmt.Errorf("no version found for GO module %q ", g.Spec.Module)
	}

	resultSource.Information = version
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("version %s found for the GO module %q", version, g.Spec.Module)

	return nil

}
