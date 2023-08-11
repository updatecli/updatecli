package dockerdigest

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source retrieves Docker image tag digest from a registry
func (ds *DockerDigest) Source(workingDir string, resultSource *result.Source) error {
	refName := ds.spec.Image
	if ds.spec.Tag != "" {
		refName += ":" + ds.spec.Tag
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

	digestWithTag := ds.spec.Tag + "@" + digest.String()
	imageDigest := ref.Context().Name() + ":" + digestWithTag

	resultSource.Result = result.SUCCESS
	resultSource.Information = digestWithTag
	resultSource.Description = fmt.Sprintf("Docker Image Tag %s resolved to digest %s",
		ref.String(), imageDigest)

	return nil
}
