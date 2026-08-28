package gomod

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Target updates a module version in a go.mod file
func (g *GoMod) Target(_ context.Context, source string, scm scm.ScmHandler, resolver utils.Resolver, dryRun bool, resultTarget *result.Target) (err error) {

	version := source
	if g.spec.Version != "" {
		version = g.spec.Version
	}

	resultTarget.NewInformation = version

	filename, err := resolver.Resolve(g.filename)
	if err != nil {
		return fmt.Errorf("invalid file path %q: %w", g.filename, err)
	}

	resultTarget.Information, resultTarget.NewInformation, resultTarget.Changed, err = g.setVersion(version, filename, dryRun)
	if err != nil {
		return err
	}

	if !resultTarget.Changed {
		switch g.kind {
		case kindGolang:
			resultTarget.Description = fmt.Sprintf("go.mod already set Golang version to %q", version)
		case kindModule:
			resultTarget.Description = fmt.Sprintf("go.mod already has Module %q set to version %q", g.spec.Module, version)
		}

		resultTarget.Result = result.SUCCESS

		return nil
	}

	resultTarget.Result = result.ATTENTION

	if dryRun {
		switch g.kind {
		case kindGolang:
			resultTarget.Description = fmt.Sprintf("go.mod should update Golang version from %q to %q",
				resultTarget.Information,
				resultTarget.NewInformation)
		case kindModule:
			resultTarget.Description = fmt.Sprintf("go.mod should update Module path %q version from %q to %q",
				g.spec.Module,
				resultTarget.Information,
				resultTarget.NewInformation)
		}

		return nil
	}

	resultTarget.Files = append(resultTarget.Files, filename)

	switch g.kind {
	case kindGolang:
		resultTarget.Description = fmt.Sprintf("go.mod updated Golang version from %q to %q",
			resultTarget.Information,
			resultTarget.NewInformation)

	case kindModule:
		resultTarget.Description = fmt.Sprintf("go.mod updated Module path %q version from %q to %q",
			g.spec.Module,
			resultTarget.Information,
			resultTarget.NewInformation)
	}

	return nil
}
