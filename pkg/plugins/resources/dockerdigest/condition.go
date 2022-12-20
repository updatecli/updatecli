package dockerdigest

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (ds *DockerDigest) Condition(source string) (bool, error) {
	refName := ds.spec.Image
	switch ds.spec.Digest == "" {
	case true:
		refName += "@" + source
	case false:
		refName += ds.spec.Digest
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return false, fmt.Errorf("invalid image %s: %w", refName, err)
	}
	_, err = remote.Head(ref, ds.options...)

	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code 404") {
			logrus.Infof("%s The Docker image %s doesn't exist.",
				result.FAILURE,
				refName,
			)
			return false, nil
		}
		return false, err
	}

	logrus.Infof("%s The Docker image %s exists and is available.",
		result.SUCCESS,
		refName,
	)

	return err == nil, err
}

func (ds *DockerDigest) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return ds.Condition(source)
}
