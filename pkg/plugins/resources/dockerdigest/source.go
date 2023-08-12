package dockerdigest

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source retrieves Docker image tag digest from a registry
func (ds *DockerDigest) Source(workingDir string, resultSource *result.Source) error {
	refTag := ""
	refName := ds.spec.Image

	if ds.spec.Tag != "" {
		if strings.HasPrefix(ds.spec.Tag, "@") {
			return fmt.Errorf("invalid tag %s: only contain a digest", ds.spec.Image)
		}

		refTagArray := strings.Split(ds.spec.Tag, "@")
		refTag = strings.TrimPrefix(refTagArray[0], ":")

		refName += ":" + refTag
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return fmt.Errorf("invalid image %s: %w", refName, err)
	}

	image, err := remote.Image(ref, ds.options...)
	if err != nil {
		return fmt.Errorf("unable to retrieve image %s: %w", refName, err)
	}

	digest, err := image.Digest()
	if err != nil {
		return fmt.Errorf("unable to retrieve image digest %s: %w", refName, err)
	}

	finalDigest := refTag + "@" + digest.String()
	imageDigest := ref.Context().Name() + ":" + finalDigest

	if ds.spec.HideTag {
		imageDigest = ref.Context().Name() + "@" + digest.String()
		finalDigest = "@" + digest.String()
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = finalDigest
	resultSource.Description = fmt.Sprintf("Docker Image Tag %s resolved to digest %s",
		ref.String(), imageDigest)

	return nil
}
