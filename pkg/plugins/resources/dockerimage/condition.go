package dockerimage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// ConditionFromSCM returns an error because it's not supported
func (di *DockerImage) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	logrus.Warningf("SCM configuration is not supported for condition of type dockerimage. Remove the `scm` directive from condition to remove this warning message")
	return di.Condition(source)
}

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (di *DockerImage) Condition(source string) (bool, error) {

	// Errors if both source input value and specified Tag are empty
	if di.image.Tag == "" && source == "" {
		return false, fmt.Errorf("condition validation error for the image %q: source input is empty and no explicit tag is specified.", di.spec.Image)
	}

	// An empty input source value means that the attribute disablesourceinput is set to true
	if source != "" {
		di.image.Tag = source
	}

	logrus.Debugf(
		"Searching digest for the image %q with the tag %q",
		di.spec.Image,
		di.spec.Tag,
	)
	digest, err := di.registry.Digest(di.image)
	if err != nil {
		return false, err
	}

	if digest == "" {
		logrus.Infof("%s The Docker image %s doesn't exist.",
			result.FAILURE,
			di.image.FullName(),
		)
		return false, nil
	}

	logrus.Infof("%s The Docker image %s exists and is available.",
		result.SUCCESS,
		di.image.FullName(),
	)

	return true, nil

}
