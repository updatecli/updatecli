package dockerdigest

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source retrieve docker image tag digest from a registry
func (ds *DockerDigest) Source(workingDir string) (string, error) {
	logrus.Debugf(
		"Searching digest for the image %q with the tag %q",
		ds.spec.Image,
		ds.spec.Tag,
	)
	digest, err := ds.registry.Digest(ds.image)
	if err != nil {
		return "", err
	}

	if digest == "" {
		return "", fmt.Errorf("%s No Digest found for the docker image %s",
			result.FAILURE,
			ds.image.FullName(),
		)
	}

	logrus.Infof("%s Digest %q found for the docker image %s.",
		result.SUCCESS,
		digest,
		ds.image.FullName(),
	)
	logrus.Infof("\tRemark: Do not forget to add @sha256 after your the docker image name")
	logrus.Infof("\tExample: %v@sha256:%v",
		ds.spec.Image,
		digest,
	)

	return digest, nil
}
