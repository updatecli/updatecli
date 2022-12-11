package dockerimage

import (
	"fmt"
)

func (di *DockerImage) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Docker Image")
}
