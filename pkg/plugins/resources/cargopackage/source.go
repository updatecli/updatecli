package cargopackage

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Source returns the latest npm package version
func (cp CargoPackage) Source(ctx context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	logrus.Debugf("Registry RootDir: %s, base directory: %s", cp.registry.RootDir, pathResolver.Dir())
	switch cp.isSCM {
	case true:
		// With an scm, the registry data is in the checkout.
		cp.registry.RootDir = pathResolver.Dir()
	case false:
		if cp.registry.RootDir != "" {
			// An empty RootDir means "no local registry checkout", so it must not be
			// resolved to the base directory.
			cp.registry.RootDir = pathResolver.Join(cp.registry.RootDir)
		}
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
