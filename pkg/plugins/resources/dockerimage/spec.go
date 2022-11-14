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
	if repository == "" {
		return ""
	}
	ref, err := name.ParseReference(repository)
	if err != nil {
		logrus.Debugf("Unable to parse repository %q: %v", repository, err)
	}
	return ref.Context().RegistryStr()
}

// NewDockerImageSpecFromImage return a new docker image specification using an image provided as parameter
func NewDockerImageSpecFromImage(image, tag string, auths map[string]docker.InlineKeyChain) *Spec {

	newVersionFilter := version.NewFilterFromValue(tag)

	dockerimagespec := Spec{
		Image: image,
	}

	switch newVersionFilter {
	case nil:
		// Option 1
		// We couldn't identify a good versionFilter so we do not return any dockerimage spec
		// At the time of writing, semantic versioning is the only way to have reliable results
		// accross the different registries.
		// More information on https://github.com/updatecli/updatecli/issues/977
		logrus.Warningf("We couldn't identify version filtering rule for container tag %q", image+":"+tag)
		return nil

		// Option 2
		// If we couldn't identify a versionFilter then we fallback to semantic versioning
		// as it's the only way to have reliable results accross the different registries
		// More information on https://github.com/updatecli/updatecli/issues/977
		//dockerimagespec.VersionFilter = version.Filter{
		//	Kind: version.SEMVERVERSIONKIND,
		//}
	default:
		dockerimagespec.VersionFilter = *newVersionFilter
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

		logrus.Debug(warningMessage)
	}

	return &dockerimagespec
}
