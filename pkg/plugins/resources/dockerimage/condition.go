package dockerimage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (di *DockerImage) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for condition of type dockerimage. Remove the `scm` directive from condition to remove this warning message")
	}

	version := source
	if di.spec.Tag != "" {
		version = di.spec.Tag
	}

	ref, err := di.createRef(version)
	if err != nil {
		return false, "", err
	}

	found := true

	if len(di.spec.Architectures) == 0 {
		found, err = di.checkImage(ref, "")
		if err != nil {
			return false, "", err
		}
	} else {
		for _, arch := range di.spec.Architectures {
			foundArchitecture, err := di.checkImage(ref, arch)
			if err != nil {
				return false, "", err
			}
			if !foundArchitecture {
				found = false
				break
			}
		}
	}

	if found {
		return true, fmt.Sprintf("docker image %s:%s found", di.spec.Image, version), nil
	}

	return false, fmt.Sprintf("docker image %s:%s not found", di.spec.Image, version), nil
}
