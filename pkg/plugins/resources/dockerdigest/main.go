package dockerdigest

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

/*
"dockerdigest" defines the specification for retrieving a container image digest from a registry.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "architecture" defines the platform of the container image, as "<architecture>" or "<os>/<architecture>[/<variant>]".
	//
	// compatible:
	//   * source
	//
	// default:
	//   linux/amd64
	//
	// remark:
	//   * when unset, the source returns the digest of the image as stored in the registry,
	//     which is the image index digest for a multi platform image.
	//   * when set, the source returns the digest of the image matching the platform.
	//   * the os defaults to "linux" when only the architecture is set.
	//
	// example:
	//   * architecture: amd64
	//   * architecture: linux/arm64
	//   * architecture: linux/arm/v7
	//
	Architecture string `yaml:",omitempty"`
	// "image" defines the container image name.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * image: updatecli/updatecli
	//   * image: ghcr.io/updatecli/updatecli
	//
	Image string `yaml:",omitempty" jsonschema:"required"`
	// "tag" defines the container image tag.
	//
	// compatible:
	//   * source
	//
	// default:
	//   latest
	//
	// remark:
	//   * a digest appended to the tag, as in "latest@sha256:...", is ignored.
	//
	// example:
	//   * tag: latest
	//   * tag: v0.1.0
	//
	Tag string `yaml:",omitempty"`
	// "digest" defines the container image digest to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	// remark:
	//   * it accepts "sha256:<digest>", "@sha256:<digest>" or "<tag>@sha256:<digest>".
	//
	// example:
	//   * digest: sha256:ce782db15ab5491c6c6178da8431b3db66988ccd11512034946a9667846952a6
	//
	Digest                string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// "hidetag" removes the tag from the source output.
	//
	// compatible:
	//   * source
	//
	// default:
	//   false
	//
	// remark:
	//   * when false, the source returns "<tag>@sha256:<digest>".
	//   * when true, the source returns "@sha256:<digest>".
	//
	HideTag bool `yaml:",omitempty"`
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

	err = newSpec.Validate()
	if err != nil {
		return nil, err
	}

	keychains := []authn.Keychain{}

	if !newSpec.Empty() {
		keychains = append(keychains, newSpec.InlineKeyChain)
	}

	keychains = append(keychains, authn.DefaultKeychain)

	os, architecture, variant := getOSArch(newSpec.Architecture)
	platform := v1.Platform{Architecture: architecture, OS: os}

	if variant != "" {
		platform.Variant = variant
	}

	newResource.options = append(newResource.options, remote.WithPlatform(platform))
	newResource.options = append(newResource.options, remote.WithAuthFromKeychain(authn.NewMultiKeychain(keychains...)))
	newResource.options = append(newResource.options, remote.WithTransport(httpclient.ProxyOnlyTransport()))
	return newResource, nil

}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (d *DockerDigest) Changelog(from, to string) *result.Changelogs {
	return nil
}

// getOSArch returns the os, architecture and variant from a string
func getOSArch(input string) (os, architecture, variant string) {

	if input == "" {
		return "linux", "amd64", ""
	}

	os = "linux"
	architecture = input
	variant = ""

	splitArchitecture := strings.Split(input, "/")

	if len(splitArchitecture) > 1 {
		os = splitArchitecture[0]
		architecture = splitArchitecture[1]
	}

	if len(splitArchitecture) > 2 {
		variant = splitArchitecture[2]
	}
	return os, architecture, variant
}

// ReportConfig returns a cleaned version of the configuration
// to identify the resource without any sensitive information or context specific data.
func (d *DockerDigest) ReportConfig() interface{} {
	return Spec{
		Image:        d.spec.Image,
		Tag:          d.spec.Tag,
		Digest:       d.spec.Digest,
		Architecture: d.spec.Architecture,
		HideTag:      d.spec.HideTag,
	}
}
