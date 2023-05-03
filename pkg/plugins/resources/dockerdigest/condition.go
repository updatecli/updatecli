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

func (ds *DockerDigest) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningln("scm is not supported, ignoring")
	}

	refName := ds.spec.Image
	switch ds.spec.Digest == "" {
	case true:
		refName += "@" + source
	case false:
		refName += ds.spec.Digest
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return fmt.Errorf("invalid image %s: %w", refName, err)
	}
	_, err = remote.Head(ref, ds.options...)

	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code 404") {

			resultCondition.Result = result.FAILURE
			resultCondition.Pass = false
			resultCondition.Description = fmt.Sprintf("the Docker image %s doesn't exist.",
				refName,
			)
			return nil
		}
		return err
	}

	resultCondition.Result = result.SUCCESS
	resultCondition.Pass = true
	resultCondition.Description = fmt.Sprintf("the Docker image %s exists and is available.",
		refName,
	)

	return nil
}
