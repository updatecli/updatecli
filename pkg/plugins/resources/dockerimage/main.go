package dockerimage

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/mitchellh/mapstructure"
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

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (di *DockerImage) Changelog() string {
	return ""
}

func (di *DockerImage) createRef(source string) (name.Reference, error) {
	refName := di.spec.Image
	switch di.spec.Tag == "" {
	case true:
		refName += ":" + source
	case false:
		refName += ":" + di.spec.Tag
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return nil, fmt.Errorf("invalid image %s: %w", refName, err)
	}

	return ref, nil
}

// checkImage checks if a container reference exists on the "remote" registry with a given set of options
func (di *DockerImage) checkImage(ref name.Reference, arch string) (bool, error) {
	var remoteOptions []remote.Option
	if arch != "" {
		remoteOptions = append(di.options, remote.WithPlatform(v1.Platform{Architecture: arch, OS: "linux"}))
	}

	descriptor, err := remote.Get(ref, remoteOptions...)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code 404") {
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
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}
