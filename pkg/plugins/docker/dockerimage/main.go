package dockerimage

import (
	"fmt"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/docker"
)

// Image represents a Docker Image to be initialized with respect to the default conventions
// as per https://docs.docker.com/engine/reference/commandline/tag/#extended-description
type Image struct {
	Registry     string
	Namespace    string
	Repository   string
	Tag          string
	Architecture string
}

// New returns an initialized Image object.
// If it fails, then the returned Image is empty and an error is returned
func New(imageFullName, architecture string) (*Image, error) {
	newImage := Image{
		Namespace:    namespace(imageFullName),
		Registry:     registry(imageFullName),
		Repository:   repository(imageFullName),
		Tag:          tag(imageFullName),
		Architecture: architecture,
	}

	if newImage.Architecture == "" {
		logrus.Warningf("No architecture specified for the image %q. Using the default %q.",
			newImage.FullName(),
			docker.DefaultImageArchitecture,
		)
		newImage.Architecture = docker.DefaultImageArchitecture
	}

	err := newImage.Validate()
	if err != nil {
		return &Image{}, err
	}

	return &newImage, nil
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (i *Image) Validate() error {
	var validationErrors []string

	if len(i.Repository) == 0 {
		validationErrors = append(validationErrors, "The specified Docker Image name is invalid: cannot detect the 'repository' part (e.g. short name).")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// FullName returns the strings (humanly readable) representation of the image
func (i *Image) FullName() string {
	return fmt.Sprintf("%s/%s/%s:%s", i.Registry, i.Namespace, i.Repository, i.Tag)
}

// tag returns the tag of the provided image (full) name
// based on the rules described at https://docs.docker.com/engine/reference/commandline/tag/#extended-description.
// The default value is defined by the constant 'docker.DefaultImageNamespace'
func tag(imageFullName string) string {
	parsedImageName := strings.Split(imageFullName, ":")
	size := len(parsedImageName)

	// If there are 2 instances of ":" found then no doubt: tag is the last one
	if size > 2 {
		return parsedImageName[size-1]
	}

	// If there is 1 instance of ":" found, it can be either a port or a tag
	// If it is a tag, there should not be ANY "/" found on the last member
	if size == 2 {
		candidateTag := parsedImageName[size-1]
		if !strings.Contains(candidateTag, "/") {
			return parsedImageName[size-1]
		}
	}

	// Otherwise return the default tag
	return docker.DefaultImageTag
}

// namespace returns the namespace of the provided image (full) name
// based on the rules described at https://docs.docker.com/engine/reference/commandline/tag/#extended-description.
// The default value is defined by the constant 'docker.DefaultImageNamespace'
func namespace(imageFullName string) string {
	// Remove tag if there is any
	imageTag := tag(imageFullName)
	parsedImageWithoutTag := strings.TrimSuffix(imageFullName, ":"+imageTag)
	parsedImageName := strings.Split(parsedImageWithoutTag, "/")

	// Either empty image name or no registry specified: return the default value
	if len(parsedImageName) < 2 {
		return docker.DefaultImageNamespace
	}

	// There is no doubt that the first element is the registry; return the 2nd member
	if len(parsedImageName) > 2 {
		return parsedImageName[1]
	}

	// Remove the ambiguity: if the first member starts with the registry hostname,
	// then there are no namespace specified: return the default
	if strings.HasPrefix(parsedImageName[0], registry(imageFullName)) {
		return docker.DefaultImageNamespace
	}

	// otherwise return the first member
	return parsedImageName[0]
}

// repository returns the "repository" part of the provided image (full) name
// based on the rules described at https://docs.docker.com/engine/reference/commandline/tag/#extended-description.
// Please note that if the empty string is returned, then it means that the provided imageFullName is invalid
func repository(imageFullName string) string {

	imageTag := tag(imageFullName)
	imageRegistry := registry(imageFullName)
	imageNamespace := namespace(imageFullName)

	// Strip the tag if it exists
	parsedImageWithoutTag := strings.TrimSuffix(imageFullName, ":"+imageTag)

	// Strip the registry if it exists
	parsedImageName := strings.Split(parsedImageWithoutTag, "/")

	if strings.HasPrefix(parsedImageName[0], imageRegistry) {
		parsedImageName = parsedImageName[1:]
	}

	// Strip the namespace if it exists
	if parsedImageName[0] == imageNamespace {
		parsedImageName = parsedImageName[1:]
	}

	return parsedImageName[0]
}

// registry returns the registry hostname from the provided Docker Image full name,
// based on the rules described at https://docs.docker.com/engine/reference/commandline/tag/#extended-description
// The default value is defined by the constant 'docker.DefaultRegistryHostname'
func registry(imageFullName string) string {
	parsedImageName := strings.Split(imageFullName, "/")

	if len(parsedImageName) < 2 {
		// Either empty image name or no registry specified: it's the default DockerHub registry
		return docker.DefaultImageRegistry
	}

	candidateRegistryHostname := parsedImageName[0]

	// Check if there is a port in the "candidate" string
	_, _, err := net.SplitHostPort(candidateRegistryHostname)
	if err == nil {
		// If there is a port then return both host and port
		return candidateRegistryHostname
	}

	// Check if the candidate resolves to an hostname (either IP of domain)
	_, err = net.LookupHost(candidateRegistryHostname)
	if err == nil {
		// If it resolves, then it's the registry hostname
		return candidateRegistryHostname
	}

	// If it's not resolved, then it's not a registry hostname
	// and we are on the Docker Hub
	return docker.DefaultImageRegistry
}
