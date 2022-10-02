package dockerdigest

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source retrieve docker image tag digest from a registry
func (ds *DockerDigest) Source(workingDir string) (string, error) {
	refName := ds.spec.Image
	if ds.spec.Tag != "" {
		refName += ":" + ds.spec.Tag
	}
	ref, err := name.ParseReference(refName)
	if err != nil {
		return "", fmt.Errorf("invalid image %s: %w", refName, err)
	}
	image, err := remote.Image(ref, ds.options...)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve image %s: %w", refName, err)
	}
	digest, err := image.Digest()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve image digest %s: %w", refName, err)
	}
	imageDigest := ref.Context().Name() + "@" + digest.String()
	logrus.Infof("%s Docker Image Tag %s resolved to digest %s", result.SUCCESS, ref.String(), imageDigest)
	return imageDigest, nil
}
