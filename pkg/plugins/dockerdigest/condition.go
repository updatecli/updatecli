package dockerdigest

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (ds *DockerDigest) Condition(source string) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin Docker Digest")
}

func (ds *DockerDigest) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return ds.Condition(source)
}
