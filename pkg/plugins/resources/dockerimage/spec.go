package dockerimage

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C][T] Architecture specifies the container image architecture such as `amd64`
	Architecture string `yaml:",omitempty"`
	// [S][C][T] Image specifies the container image such as `updatecli/updatecli`
	Image string `yaml:",omitempty"`
	// [C][T] Tag specifies the container image tag such as `latest`
	Tag                   string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

func sanitizeRegistryEndpoint(repository string) string {
	ref, err := name.ParseReference(repository)
	if err != nil {
		logrus.Debugf("Unable to parse repository %q: %v", repository, err)
	}
	return ref.Context().RegistryStr()
}

// NewDockerImageSpecFromImage return a new docker image specification using an image provided as parameter
func NewDockerImageSpecFromImage(image string, auths map[string]docker.InlineKeyChain) Spec {
	dockerimagespec := Spec{
		Image: image,
		VersionFilter: version.Filter{
			Kind: version.SEMVERVERSIONKIND,
		},
	}

	registry := sanitizeRegistryEndpoint(image)

	credential, found := auths[registry]

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

		for endpoint := range auths {
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
