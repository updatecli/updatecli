package tag

import (
	"fmt"
)

// Target ensure that a specific release exist on gitea, otherwise creates it
func (g *Gitea) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("target not supported for the plugin Gitea Tags")
}
