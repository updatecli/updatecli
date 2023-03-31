package tag

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target ensure that a specific release exist on bitbucket, otherwise creates it
func (g *Stash) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("target not supported for the plugin Bitbucket Tags")
}

func (g Stash) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("target not supported for the plugin Bitbucket Tags")
}
