package dockerimage

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C][T] Architecture specifies the container image architecture such as `amd64`
	Architecture string `yaml:",omitempty"`
	// [C] Architectures specifies a list of architectures to check container images for (conditions only)
	Architectures []string `yaml:",omitempty"`
	// [S][C][T] Image specifies the container image such as `updatecli/updatecli`
	Image string `yaml:",omitempty"`
	// [C][T] Tag specifies the container image tag such as `latest`
	Tag                   string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// DockerImage defines a resource of type "dockerimage"
type DockerImage struct {
	spec    Spec
	options []remote.Option
	// versionFilter holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	foundVersion  version.Version
}

// New returns a reference to a newly initialized DockerImage object from a dockerimage.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*DockerImage, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if len(newSpec.Architectures) > 0 {
		if newSpec.Architecture != "" {
			return nil, fmt.Errorf("Validation error in the resource of type 'dockerimage': the attributes `spec.architecture` and `spec.architecture` are mutually exclusive")
		}

	} else {
		if newSpec.Architecture == "" {
			newSpec.Architecture = "amd64"
		}

		// Move the "single" architecture to the "multiple" (used everywhere) and discard it
		newSpec.Architectures = append(newSpec.Architectures, newSpec.Architecture)
		newSpec.Architecture = ""
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	newResource := &DockerImage{
		spec:          newSpec,
		versionFilter: newFilter,
	}

	err = newSpec.InlineKeyChain.Validate()
	if err != nil {
		return nil, err
	}

	keychains := []authn.Keychain{}

	if !newSpec.InlineKeyChain.Empty() {
		keychains = append(keychains, newSpec.InlineKeyChain)
	}

	keychains = append(keychains, authn.DefaultKeychain)

	newResource.options = append(newResource.options, remote.WithAuthFromKeychain(authn.NewMultiKeychain(keychains...)))

	if len(newSpec.Architectures) == 1 {
		newResource.options = append(newResource.options, remote.WithPlatform(v1.Platform{Architecture: newSpec.Architectures[0], OS: "linux"}))
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (di *DockerImage) Changelog() string {
	return ""
}
