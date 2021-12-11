package dockerdigest

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerregistry"
)

// Spec defines a specification for a "dockerDigest" resource
// parsed from an updatecli manifest file
type Spec struct {
	Architecture string
	Image        string
	Tag          string
	Username     string
	Password     string
	Token        string
}

// DockerDigest defines a resource of kind "dockerDigest" to interact with a docker registry
type DockerDigest struct {
	spec     Spec
	registry dockerregistry.Registry
	image    dockerimage.Image
}

// New returns a reference to a newly initialized DockerDigest object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*DockerDigest, error) {
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

	newResource := &DockerDigest{
		spec:     newSpec,
		registry: newRegistry,
		image:    newImage,
	}

	err = newResource.Validate()
	if err != nil {
		return nil, err
	}

	return newResource, nil
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (ds *DockerDigest) Validate() error {
	var validationErrors []string

	if len(ds.spec.Token) > 0 {
		if len(ds.spec.Username) > 0 || len(ds.spec.Password) > 0 {
			validationErrors = append(validationErrors, "Specifying a (bearer) token is invalid when a username and a password are provided.")
		}
	}

	if len(ds.spec.Username) > 0 && len(ds.spec.Password) == 0 {
		validationErrors = append(validationErrors, "Docker registry username provided but not the password")
	} else if len(ds.spec.Username) == 0 && len(ds.spec.Password) > 0 {
		validationErrors = append(validationErrors, "Docker registry password provided but not the username")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}
