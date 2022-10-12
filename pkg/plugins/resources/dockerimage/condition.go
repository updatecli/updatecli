package dockerimage

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
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
	refName := di.spec.Image
	if di.spec.Tag != "" {
		refName += ":" + di.spec.Tag
	}
	ref, err := name.ParseReference(refName)
	if err != nil {
		return false, fmt.Errorf("invalid image %s: %w", refName, err)
	}
	_, err = remote.Head(ref, di.options...)
	return err == nil, err
}
