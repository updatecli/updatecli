package gomod

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target is not supported for the Golang resource
func (l *GoMod) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Go")
}

// TargetFromSCM is not supported for the Golang resource
func (l *GoMod) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Go")
}
