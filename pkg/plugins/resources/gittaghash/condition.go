package gittaghash

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (gth GitTagHash) Condition(source string) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin Git Tag Hash, use Git Tag")
}

func (gth GitTagHash) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin Git Tag Hash, use Git Tag")
}
