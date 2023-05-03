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

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (di *DockerImage) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for condition of type dockerimage. Remove the `scm` directive from condition to remove this warning message")
	}

	refName := di.spec.Image
	switch di.spec.Tag == "" {
	case true:
		refName += ":" + source
	case false:
		refName += ":" + di.spec.Tag
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return fmt.Errorf("invalid image %s: %w", refName, err)
	}

	found := false
	switch len(di.spec.Architectures) {
	case 0:
		// If only one architecture is specified, then di.options are already set: let's proceed
		found, err = checkImage(ref, di.spec.Architecture, di.options)
		if err != nil {
			return err
		}

	default:
		for _, arch := range di.spec.Architectures {
			remoteOptions := append(di.options, remote.WithPlatform(v1.Platform{Architecture: arch, OS: "linux"}))

			found, err = checkImage(ref, arch, remoteOptions)
			if err != nil {
				return err
			}
			if found {
				break
			}
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
