package dockerfile

import (
	"fmt"
)

func (df *Dockerfile) Source(workingDir string) (string, error) {
	return "", fmt.Errorf("Source is not supported for the plugin 'dockerfile'")
}
