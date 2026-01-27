package dockerdigest

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// Spec defines a specification for a "dockerdigest" resource parsed from an updatecli manifest file
type Spec struct {
	// architecture specifies the container image architecture such as `amd64`
	//
	// compatible:
	// 	* source
	// 	* condition
	//
	// default: amd64
	Architecture string `yaml:",omitempty"`
	// image specifies the container image such as `updatecli/updatecli`
	//
	// example: `updatecli/updatecli`
	//
	// compatible:
	// 	* source
	// 	* condition
	Image string `yaml:",omitempty" jsonschema:"required"`
	// tag specifies the container image tag such as `latest`
	//
	// compatible:
	// 	* source
	// 	* condition
	Tag string `yaml:",omitempty"`
	// digest specifies the container image digest such as `sha256:ce782db15ab5491c6c6178da8431b3db66988ccd11512034946a9667846952a6`
	//
	// compatible:
	// 	* condition
	//
	// default:
	// 	When used from a condition, the default value is set to the linked source output.
	Digest                string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
	// hideTag specifies if the tag should be hidden from the digest
	//
	// compatible:
	// 	* source
	//
	// default:
	// 	false
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
	}
}
