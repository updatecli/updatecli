package dockercompose

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func sanitizeRegistryEndpoint(repository string) string {
	ref, err := name.ParseReference(repository)
	if err != nil {
		logrus.Debugf("Unable to parse repository %q: %v", repository, err)
	}
	return ref.Context().RegistryStr()
}

func (h DockerCompose) generateSourceDockerImageSpec(image string) dockerimage.Spec {
	dockerimagespec := dockerimage.Spec{
		Image: image,
		VersionFilter: version.Filter{
			Kind: version.SEMVERVERSIONKIND,
		},
	}

	registry := sanitizeRegistryEndpoint(image)

	credential, found := h.spec.Auths[registry]

	switch found {
	case true:
		if credential.Password != "" {
			dockerimagespec.Password = credential.Password
		}
		if credential.Token != "" {
			dockerimagespec.Token = credential.Token
		}
		if credential.Username != "" {
			dockerimagespec.Username = credential.Username
		}
	default:

		registryAuths := []string{}

		for endpoint := range h.spec.Auths {
			logrus.Printf("Endpoint:\t%q\n", endpoint)
			registryAuths = append(registryAuths, endpoint)
		}

		warningMessage := fmt.Sprintf(
			"no credentials found for docker registry %q hosting image %q, among %q",
			registry,
			image,
			strings.Join(registryAuths, ","))

		if len(registryAuths) == 0 {
			warningMessage = fmt.Sprintf("no credentials found for docker registry %q hosting image %q",
				registry,
				image)
		}

		logrus.Warning(warningMessage)
	}

	return dockerimagespec
}
