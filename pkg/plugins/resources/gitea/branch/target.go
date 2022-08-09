package branch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target ensure that a specific release exist on gitea, otherwise creates it
func (g *Gitea) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("target not supported for the plugin Gitea branch")

}

func (g Gitea) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("target not supported for the plugin GitHub Release")
}
