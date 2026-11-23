package pypi

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Source returns the latest matching PyPI package version.
func (p *Pypi) Source(ctx context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	ver, err := p.getVersions(ctx)
	if err != nil {
		/*
			The age filter discarded every version the versionfilter matches, or couldn't
			check them, which is an expected state rather than a failure, so the source is
			skipped instead. The error tells why, such as a registry reporting no upload time.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = fmt.Sprintf("skipping the PyPI package %q: %s", p.spec.Name, err)
			return nil
		}

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
