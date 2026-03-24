package npm

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest npm package version
func (n Npm) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {
	version, _, err := n.getVersions(ctx)
	if err != nil {
		return err
	}

	if version == "" {
		return fmt.Errorf("unknown version %s found for package name %s ", version, n.spec.Name)
	}

	resultSource.Information = version
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("version %s found for package name %q", version, n.spec.Name)

	return nil

}
