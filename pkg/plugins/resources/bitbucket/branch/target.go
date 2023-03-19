package branch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target ensure that a specific release exist on bitbucket, otherwise creates it
func (g *Bitbucket) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("target not supported for the plugin Bitbucket branch")

}

func (g Bitbucket) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("target not supported for the plugin GitHub Release")
}
