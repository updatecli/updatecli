package dockerimage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (di *DockerImage) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for condition of type dockerimage. Remove the `scm` directive from condition to remove this warning message")
	}

	ref, err := di.createRef(source)
	if err != nil {
		return err
	}

	found := true

	for _, arch := range di.spec.Architectures {
		foundArchitecture, err := di.checkImage(ref, arch)
		if err != nil {
			return err
		}
		if !foundArchitecture {
			found = false
			break
		}
	}

	if found {
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("docker image %q:%q found", di.spec.Image, source)
		return nil
	}

	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE
	resultCondition.Description = fmt.Sprintf("docker image %q:%q not found", di.spec.Image, source)

	return nil
}
