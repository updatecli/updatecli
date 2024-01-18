package language

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (g *Language) Source(workingDir string, resultSource *result.Source) error {
	_, err := g.versions()
	if err != nil {
		return fmt.Errorf("retrieving golang version: %w", err)
	}

	resultSource.Information = g.Version.GetVersion()

	if resultSource.Information == "" {
		return fmt.Errorf("no Golang version found matching pattern %q",
			g.Spec.VersionFilter.Pattern,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Golang version %s found",
		g.Version.GetVersion(),
	)

	return nil

}
