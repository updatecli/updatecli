package dockerimage

import (
	"fmt"
)

func (di *DockerImage) Source(workingDir string) (string, error) {
	return "", fmt.Errorf("Source is not supported for the plugin Docker Image")
}
