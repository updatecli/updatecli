package dockerimage

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerregistry"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	Architecture string
	Image        string
	Tag          string
	Username     string
	Password     string
	Token        string
}

// DockerImage defines a resource of type "dockerimage"
type DockerImage struct {
	spec     Spec
	registry dockerregistry.Registry
	image    dockerimage.Image
}

// New returns a reference to a newly initialized DockerImage object from a dockerimage.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*DockerImage, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newImage, err := dockerimage.New(
		// concatenate user-provided image name (that may contain registry and namespace along with the namespace) with the user-provided tag
		fmt.Sprintf("%s:%s", newSpec.Image, newSpec.Tag),
		newSpec.Architecture,
	)
	if err != nil {
		return nil, err
	}

	newRegistry := dockerregistry.New(
		newImage.Registry,
		newSpec.Username,
		newSpec.Password,
	)

	newResource := &DockerImage{
		spec:     newSpec,
		registry: newRegistry,
		image:    *newImage,
	}

	err = newResource.Validate()
	if err != nil {
		return nil, err
	}

	return newResource, nil
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (di *DockerImage) Validate() error {
	var validationErrors []string

	if len(di.spec.Token) > 0 {
		if len(di.spec.Username) > 0 && len(di.spec.Password) > 0 {
			validationErrors = append(validationErrors, "Specifying a (bearer) token is invalid when a username and a password are provided.")
		}
	}

	if len(di.spec.Username) > 0 && len(di.spec.Password) == 0 {
		validationErrors = append(validationErrors, "Docker registry username provided but not the password")
	} else if len(di.spec.Username) == 0 && len(di.spec.Password) > 0 {
		validationErrors = append(validationErrors, "Docker registry password provided but not the username")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (di *DockerImage) Changelog() string {
	return ""
}
