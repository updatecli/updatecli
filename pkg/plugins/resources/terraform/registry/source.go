package registry

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest version
func (t *TerraformRegistry) Source(workingDir string, resultSource *result.Source) error {
	_, err := t.versions()
	if err != nil {
		return fmt.Errorf("%s retrieving terraform registry version: %w", result.FAILURE, err)
	}

	resultSource.Information = t.Version.GetVersion()

	if resultSource.Information == "" {
		return fmt.Errorf("%s no terraform registry version matching pattern: %q", result.FAILURE, t.Spec.VersionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Terraform registry version %s found",
		t.Version.GetVersion(),
	)

	return nil
}
