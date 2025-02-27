package language

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (l *Language) Source(workingDir string, resultSource *result.Source) error {
	_, err := l.versions()
	if err != nil {
		return fmt.Errorf("retrieving golang version: %w", err)
	}

	resultSource.Information = l.Version.GetVersion()

	if resultSource.Information == "" {
		return fmt.Errorf("no Golang version found matching pattern %q",
			l.Spec.VersionFilter.Pattern,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Golang version %s found",
		l.Version.GetVersion(),
	)

	return nil

}
