package cargopackage

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (cp *CargoPackage) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Cargo Package")
}

func (cp *CargoPackage) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Cargo Package")
}
