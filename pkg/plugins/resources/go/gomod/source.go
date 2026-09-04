package gomod

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Source returns the latest go module version
func (g *GoMod) Source(_ context.Context, resolver utils.Resolver, resultSource *result.Source) error {
	filename, err := resolver.Resolve(g.filename)
	if err != nil {
		return fmt.Errorf("invalid file path %q: %w", g.filename, err)
	}

	g.foundVersion, err = g.version(filename)
	if err != nil {
		return fmt.Errorf("searching version: %w", err)
	}

	if g.foundVersion == "" {
		return fmt.Errorf("no version found for module path %q", g.spec.Module)
	}

	resultSource.Information = g.foundVersion
	resultSource.Result = result.SUCCESS

	switch g.kind {
	case kindGolang:
		resultSource.Description = fmt.Sprintf("Golang Version %s found", g.foundVersion)
	case kindModule:
		resultSource.Description = fmt.Sprintf("version %s found for GO module %q",
			g.foundVersion,
			g.spec.Module,
		)
	}

	return nil
}
