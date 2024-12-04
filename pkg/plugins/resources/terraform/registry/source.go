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

	version := t.Version.GetVersion()
	if version == "" {
		return fmt.Errorf("%s no terraform registry version matching pattern: %q", result.FAILURE, t.Spec.VersionFilter.Pattern)
	}
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: version,
	}}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Terraform registry version %s found", version)

	return nil
}
