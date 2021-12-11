package dockerimage

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (di *DockerImage) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Docker Image")
}

func (di *DockerImage) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Docker Image")
}
