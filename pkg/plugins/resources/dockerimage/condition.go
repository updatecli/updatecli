package dockerimage

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
	refName := di.spec.Image
	switch di.spec.Tag == "" {
	case true:
		refName += ":" + source
	case false:
		refName += ":" + di.spec.Tag
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return false, fmt.Errorf("invalid image %s: %w", refName, err)
	}

	if len(di.spec.Architectures) == 0 {
		// If only one architecture is specified, then di.options are already set: let's proceed
		return checkImage(ref, di.spec.Architecture, di.options)
	}

	for _, arch := range di.spec.Architectures {
		remoteOptions := append(di.options, remote.WithPlatform(v1.Platform{Architecture: arch, OS: "linux"}))

		found, err := checkImage(ref, arch, remoteOptions)
		if !found || err != nil {
			return false, err
		}
	}

	return true, err
}

// checkImage checks if a container reference exists on the "remote" registry with a given set of options
func checkImage(ref name.Reference, arch string, remoteOpts []remote.Option) (bool, error) {
	_, err := remote.Head(ref, remoteOpts...)

	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code 404") {
			logrus.Infof("%s The Docker image %s (%s) doesn't exist.",
				result.FAILURE,
				ref.Name(),
				arch,
			)
			return false, nil
		}
		return false, err
	}

	logrus.Infof("%s The Docker image %s (%s) exists and is available.",
		result.SUCCESS,
		ref.Name(),
		arch,
	)

	return true, nil
}
