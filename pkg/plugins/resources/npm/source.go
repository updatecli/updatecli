package npm

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// Source returns the latest npm package version
func (n Npm) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {
	version, _, err := n.getVersions(ctx)
	if err != nil {
		/*
			Every published version is still cooling down, which is an expected state of
			the age filter rather than a failure, so the source is skipped instead.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = fmt.Sprintf("no version of the npm package %q matches the age filter yet", n.spec.Name)
			return nil
		}

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
