package pypi

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest matching PyPI package version.
func (p *Pypi) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {
	ver, _, err := p.getVersions(ctx)
	if err != nil {
		return err
	}

	if ver == "" {
		return fmt.Errorf("no version found for package %q", p.spec.Name)
	}

	resultSource.Information = ver
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("version %q found for PyPI package %q", ver, p.spec.Name)

	return nil
}
