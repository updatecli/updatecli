package dockerfile

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (df *Dockerfile) Source(workingDir string, resultSource *result.Source) error {
	return fmt.Errorf("Source is not supported for the plugin 'dockerfile'")
}
