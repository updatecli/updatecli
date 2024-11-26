package cargopackage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest npm package version
func (cp CargoPackage) Source(workingDir string, resultSource *result.Source) error {
	logrus.Debugf("Registry RootDir: %s, workingDir: %s", cp.registry.RootDir, workingDir)
	if cp.isSCM {
		// We are in a scm context, workingDir is holding the data
		cp.registry.RootDir = workingDir
	}

	version, _, err := cp.getVersions()
	if err != nil {
		return fmt.Errorf("get cargo packages versions: %w", err)
	}

	if version == "" {
		return fmt.Errorf("no version found for cargo package name %q", cp.spec.Package)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: version,
	}}
	resultSource.Description = fmt.Sprintf("version %q found for cargo package name %q", version, cp.spec.Package)
	return nil
}
