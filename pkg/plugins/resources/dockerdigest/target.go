package dockerdigest

import (
	"fmt"
)

func (ds *DockerDigest) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Docker Digest")
}
