package dockerdigest

import (
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// Spec defines a specification for a "dockerdigest" resource parsed from an updatecli manifest file
type Spec struct {
	// [s][c] Architecture specifies the container image architecture such as `amd64`
	Architecture string `yaml:",omitempty"`
	// [s][c] Image specifies the container image such as `updatecli/updatecli`
	Image string `yaml:",omitempty"`
	// [s] Tag specifies the container image tag such as `latest`
	Tag string `yaml:",omitempty"`
	// [c] Digest specifies the container image digest such as `@sha256:ce782db15ab5491c6c6178da8431b3db66988ccd11512034946a9667846952a6`
	Digest                string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
}

// DockerDigest defines a resource of kind "dockerDigest" to interact with a docker registry
type DockerDigest struct {
	spec    Spec
	options []remote.Option
}

// New returns a reference to a newly initialized DockerDigest object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*DockerDigest, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &DockerDigest{
		spec: newSpec,
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
	arch := newSpec.Architecture
	if arch == "" {
		arch = "amd64"
	}
	newResource.options = append(newResource.options, remote.WithPlatform(v1.Platform{Architecture: arch, OS: "linux"}))
	newResource.options = append(newResource.options, remote.WithAuthFromKeychain(authn.NewMultiKeychain(keychains...)))
	return newResource, nil

}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (d *DockerDigest) Changelog() string {
	return ""
}
