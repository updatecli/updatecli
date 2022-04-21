package dockerregistry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
)

// Registry is an interface to any Registry-specific metadata retriever (e.g. to retrieve Docker images metadatas)
type Registry interface {
	Digest(image dockerimage.Image) (string, error)
}

// RegistryAuth holds the authentication element of a given registry
type RegistryAuth struct {
	Username string
	Password string
	Token    string
}

// RegistryEndpoints holds the URL endpoints for a given registry
type RegistryEndpoints struct {
	ApiService    string
	TokenService  string
	ScopedService string
}

// DockerGenericRegistry is the main implementation of the Registry interface.
// It supports any v2 Registry API such as DockerHub, GHCR, etc.
type DockerGenericRegistry struct {
	Auth      RegistryAuth
	WebClient httpclient.HTTPClient
}

// New returns a newly initialized Registry object.
func New(hostname, username, password string) Registry {

	// Returns the initialized object
	return DockerGenericRegistry{
		Auth: RegistryAuth{
			Username: username,
			Password: password,
		},
		WebClient: http.DefaultClient,
	}
}

// Digest retrieves the digest of the provided docker image from the registry
func (dgr DockerGenericRegistry) Digest(image dockerimage.Image) (string, error) {
	endpoints := registryEndpoints(image)

	URL := fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s",
		endpoints.ApiService,
		image.Namespace,
		image.Repository,
		image.Tag,
	)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	// Provide the registry with a list of content types that we support:
	//  * Docker Registry V2 manifest list
	//  * List of OCI manifests (multiple architectures provided by the registry)
	//  * Standalone OCI manifest (only one architecture provided by the registry)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")

	// Retrieve a bearer token to authenticate further requests
	token, err := dgr.login(image)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("Could not retrieve a bearer token (empty value).")
	}

	logrus.Debugf("Bearer token retrieved successfully.")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	logrus.Debugf("Emitting a request to: %v", URL)

	res, err := dgr.WebClient.Do(req)
	if err != nil {
		return "", err
	}

	logrus.Debugf("Received the following response from the registry %q: %v", image.Registry, res)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			// Return an empty string but no error (it's the caller responsibility to handle the error)
			return "", nil
		}
		err = fmt.Errorf("Unexpected error from the registry %s for image %s:%s",
			image.Registry,
			image.Repository,
			image.Tag,
		)
		logrus.Error(err)
		return "", err
	}

	switch res.Header.Get("content-type") {
	// Standalone OCI manifest (only one architecture provided by the registry)
	// OCI compatibility matrix
	// https://github.com/opencontainers/image-spec/blob/v1.0.1/media-types.md#applicationvndociimageindexv1json
	case "application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json":
		// Note that there are no check against the image's architecture
		// since the image architecture is stored in the configuration layer
		// and it would require another HTTP request to fetch it.
		return strings.TrimPrefix(res.Header.Get("Docker-Content-Digest"), "sha256:"), nil
	// Newer registries that are OCI compliant can return a list of OCI
	// manifests (a container image is supplied for multiple architectures).
	// This format is backward compatible with the Docker Registry V2.

	case "application/vnd.oci.image.index.v1+json":
		fallthrough
	// Standard Registry v2 API (nominal case) such as DockerHub or GHCR

	case "application/vnd.docker.distribution.manifest.list.v2+json":
		type response struct {
			Manifests []struct {
				Digest   string
				Platform struct {
					Architecture string
					Os           string
				}
			}
		}

		data := response{}

		err = json.Unmarshal(body, &data)
		if err != nil {
			return "", err
		}

		for _, returnedImages := range data.Manifests {
			if returnedImages.Platform.Architecture == image.Architecture {
				digest := strings.TrimPrefix(returnedImages.Digest, "sha256:")
				return digest, nil
			}
		}

	default:
		logrus.Debugf("Returned response body:\n%v", string(body)) // Shows answer in debug mode to help diagnostics
	}

	// Unsupported or deprecated answer's type:
	// - For instance "application/vnd.docker.distribution.manifest.v1+prettyjws" means a "registry API v1" which is not used - https://docs.docker.com/registry/spec/deprecated-schema-v1
	return "", fmt.Errorf("Unsupported response type from the registry %s: Content-Type %q is either deprecated or unknown by updatecli.", image.Registry, res.Header.Get("content-type"))
}

// login returns a bearer token to authenticate further request to the docker registry
func (dgr DockerGenericRegistry) login(image dockerimage.Image) (string, error) {
	endpoints := registryEndpoints(image)

	URL := fmt.Sprintf("https://%s?scope=repository:%s/%s:pull&service=%s",
		endpoints.TokenService,
		image.Namespace,
		image.Repository,
		endpoints.ScopedService,
	)

	logrus.Debugf("Retrieving a bearer token from %s with the username %s and the associated password.",
		URL,
		dgr.Auth.Username,
	)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	// Even without username or password, DockerHub needs a token:
	// an anonymous token can be generated (https://docs.docker.com/docker-hub/download-rate-limit/)
	if dgr.Auth.Username != "" && dgr.Auth.Password != "" {
		req.SetBasicAuth(dgr.Auth.Username, dgr.Auth.Password)
	} else {
		logrus.Warningf("No username and/or password specified: trying to retrieve a token as anonymous user might not be supported or could impact the rate limiting on your network.")
	}

	logrus.Debugf("Emitting a request to %s", URL)
	res, err := dgr.WebClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("Got an HTTP response code %d while logging in to the registry %s.",
			res.StatusCode,
			image.Registry,
		)
	}

	type response struct {
		Token string
	}

	data := response{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	return data.Token, nil
}

// registryEndpoints returns a RegistryEndpoints with the endpoints (without schemes) of the registry of the provided image
func registryEndpoints(image dockerimage.Image) RegistryEndpoints {

	// DockerHub (multiple aliases)
	if strings.HasSuffix(image.Registry, "docker.io") || strings.HasSuffix(image.Registry, "hub.docker.com") {
		return RegistryEndpoints{
			ApiService:    "index.docker.io",
			TokenService:  "auth.docker.io/token",
			ScopedService: "registry.docker.io",
		}
	}

	// Quay.io (no aliases)
	if strings.HasSuffix(image.Registry, "quay.io") {
		return RegistryEndpoints{
			ApiService:    "quay.io",
			TokenService:  "quay.io/v2/auth",
			ScopedService: "quay.io",
		}
	}

	// Default: keep the same registry hostname for all endpoints
	return RegistryEndpoints{
		ApiService:    image.Registry,
		TokenService:  image.Registry + "/token",
		ScopedService: image.Registry,
	}
}
