package cargopackage

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Source returns the latest npm package version
func (cp CargoPackage) Source(ctx context.Context, resolver utils.Resolver, resultSource *result.Source) error {
	logrus.Debugf("Registry RootDir: %s, base directory: %s", cp.registry.RootDir, resolver.Dir())
	switch cp.isSCM {
	case true:
		// We are in a scm context, the base directory is holding the data
		cp.registry.RootDir = resolver.Dir()
	case false:
		cp.registry.RootDir = resolver.Join(cp.registry.RootDir)
	}

	version, _, err := cp.getVersions(ctx)
	if err != nil {
		return fmt.Errorf("get cargo packages versions: %w", err)
	}

	if version == "" {
		return fmt.Errorf("no version found for cargo package name %q", cp.spec.Package)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = version
	resultSource.Description = fmt.Sprintf("version %q found for cargo package name %q", version, cp.spec.Package)
	return nil
}
