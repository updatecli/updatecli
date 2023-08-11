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

// Condition checks if a Docker image tag digest exists in a registry
func (ds *DockerDigest) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningln("scm is not supported, ignoring")
	}

	refName := ds.spec.Image
	switch ds.spec.Digest == "" {
	case true:
		refName = joinImageTagWithName(refName, source)
	case false:
		refName = joinImageTagWithName(refName, ds.spec.Digest)
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

// joinImageTagWithName handles the different kind of join tag
func joinImageTagWithName(image, tag string) string {
	/*
		if tag start with @sha256, we assume that the tag is a digest
		if tag start with ':', we assume that the tag is a tag and not a digest
	*/
	if strings.HasPrefix(tag, ":") || strings.HasPrefix(tag, "@sha256") {
		return image + tag
	}

	// if the tag doesnt' start with @sha256 then we assume the first part is a tag
	if strings.Contains(tag, "@sha256") && !strings.HasPrefix(tag, "@sha256") {
		return image + ":" + strings.TrimPrefix(tag, ":")
	}

	return image + "@" + tag
}
