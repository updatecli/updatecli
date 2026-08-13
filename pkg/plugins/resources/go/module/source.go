package gomodule

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (g *GoModule) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {
	version, _, err := g.versions(ctx)
	if err != nil {
		/*
			Every published version is still cooling down, which is an expected state of
			the age filter rather than a failure, so the source is skipped instead.
		*/
		if errors.Is(err, ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = fmt.Sprintf("no version of the GO module %q matches the age filter yet", g.Spec.Module)
			return nil
		}

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
