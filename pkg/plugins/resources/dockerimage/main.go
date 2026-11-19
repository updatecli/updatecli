package dockerimage

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

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
			return nil, fmt.Errorf("validation error in the resource of type 'dockerimage': the attributes `spec.architecture` and `spec.architectures` are mutually exclusive")
		}
	} else {
		if newSpec.Architecture != "" {
			// Move the "single" architecture to the "multiple" (used everywhere) and discard it
			newSpec.Architectures = append(newSpec.Architectures, newSpec.Architecture)
			newSpec.Architecture = ""
		}
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	newResource := &DockerImage{
		spec:          newSpec,
		versionFilter: newFilter,
	}

	err = newSpec.Validate()
	if err != nil {
		return nil, err
	}

	if err := newSpec.Age.Validate(); err != nil {
		return nil, fmt.Errorf("wrong age spec: %w", err)
	}

	keychains := []authn.Keychain{}

	if !newSpec.Empty() {
		keychains = append(keychains, newSpec.InlineKeyChain)
	}

	keychains = append(keychains, authn.DefaultKeychain)

	newResource.options = append(newResource.options, remote.WithAuthFromKeychain(authn.NewMultiKeychain(keychains...)))
	newResource.options = append(newResource.options, remote.WithTransport(httpclient.ProxyOnlyTransport()))

	return newResource, nil
}

func (di *DockerImage) createRef(source string) (name.Reference, error) {
	refName := di.spec.Image
	refName += ":" + source

	ref, err := name.ParseReference(refName)
	if err != nil {
		return nil, fmt.Errorf("invalid image %s: %w", refName, err)
	}

	return ref, nil
}

// remoteOptions returns the registry options, restricted to a platform when arch is set.
func (di *DockerImage) remoteOptions(arch string) []remote.Option {
	remoteOptions := slices.Clone(di.options)

	if arch == "" {
		return remoteOptions
	}

	return append(remoteOptions, remote.WithPlatform(parsePlatform(arch)))
}

// parsePlatform parses a platform written "<architecture>" or "<os>/<architecture>[/<variant>]".
func parsePlatform(arch string) v1.Platform {
	platform := v1.Platform{OS: "linux", Architecture: arch}

	splitArchitecture := strings.Split(arch, "/")

	if len(splitArchitecture) > 1 {
		platform.OS = splitArchitecture[0]
		platform.Architecture = splitArchitecture[1]
	}

	if len(splitArchitecture) > 2 {
		platform.Variant = splitArchitecture[2]
	}

	return platform
}

// checkImage checks if a container reference exists on the "remote" registry with a given set of options
func (di *DockerImage) checkImage(ref name.Reference, arch string) (bool, error) {
	var queriedPlatform string

	if arch != "" {
		platform := parsePlatform(arch)
		queriedPlatform = platform.String()

		log.Debugf("Querying docker image %q, os: %q, arch: %q, variant %q", ref.Name(), platform.OS, platform.Architecture, platform.Variant)
	}

	descriptor, err := remote.Get(ref, di.remoteOptions(arch)...)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			return false, nil
		}
		return false, err
	}

	if arch != "" {

		if descriptor.MediaType == types.DockerManifestSchema1 || descriptor.MediaType == types.DockerManifestSchema1Signed {
			return false, fmt.Errorf("architecture check not supported with MediaType %q", descriptor.MediaType)
		}

		_, err = descriptor.Image()
		if err != nil {
			if strings.Contains(err.Error(), "no child with platform") {
				log.Infof("The Docker image %s (%s) doesn't exist.",
					ref.Name(),
					queriedPlatform,
				)
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}

// ReportConfig returns a new configuration with only the necessary configuration fields
// to identify the resource without any sensitive information or context specific data.
func (di *DockerImage) ReportConfig() interface{} {
	return Spec{
		Image:         di.spec.Image,
		Architectures: di.spec.Architectures,
		Architecture:  di.spec.Architecture,
		Tag:           di.spec.Tag,
		TagFilter:     di.spec.TagFilter,
		VersionFilter: di.spec.VersionFilter,
		Age:           di.spec.Age,
	}
}

// isStatus returns true if err is a registry response with the given status code.
func isStatus(err error, code int) bool {
	var transportErr *transport.Error
	return errors.As(err, &transportErr) && transportErr.StatusCode == code
}
