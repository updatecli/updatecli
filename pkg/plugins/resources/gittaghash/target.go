package gittaghash

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target creates a tag if needed from a local git repository, without pushing the tag
func (ghr GitTagHash) Target(source string, dryRun bool) (bool, error) {
	return false, fmt.Errorf("Target not supported for the plugin Git Tag Hash, use Git Tag")
}

func (ghr GitTagHash) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	return false, []string{}, "", fmt.Errorf("Target not supported for the plugin Git Tag Hash, use Git Tag")
}
