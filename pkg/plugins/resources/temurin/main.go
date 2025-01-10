package temurin

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

const temurinApiUrl string = "https://api.adoptium.net/v3"

// Http defines a resource of type "temurin"
type Temurin struct {
	spec                    Spec
	apiWebClient            httpclient.HTTPClient
	apiWebRedirectionClient httpclient.HTTPClient
	foundVersion            string
}

/*
*
New returns a reference to a newly initialized Temurin resource
or an error if the provided Spec triggers a validation error.
*
*/
func New(spec interface{}) (*Temurin, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if len(newSpec.Platforms) > 0 {
		if newSpec.Architecture != "" {
			return nil, fmt.Errorf("[temurin] `spec.platform` and `spec.architecture` are mutually exclusive.")
		}
		if newSpec.OperatingSystem != "" {
			return nil, fmt.Errorf("[temurin] `spec.platform` and `spec.operatingsystem` are mutually exclusive.")
		}
	}

	// Set defaults
	if newSpec.Result == "" {
		newSpec.Result = "version"
	}
	if newSpec.Architecture == "" {
		newSpec.Architecture = "x64"
	}
	if newSpec.Project == "" {
		newSpec.Project = "jdk"
	}
	if newSpec.OperatingSystem == "" {
		newSpec.OperatingSystem = "linux"
	}
	if newSpec.ImageType == "" {
		newSpec.ImageType = "jdk"
	}
	if newSpec.ReleaseType == "" {
		newSpec.ReleaseType = "ga"
	}
	if newSpec.SpecificVersion != "" && newSpec.FeatureVersion != 0 {
		return nil, fmt.Errorf("[temurin] resource with both 'specificversion' and 'featureversion' specified which are mutually exclusive.")
	}

	httpClient := httpclient.NewRetryClient().(*http.Client)

	newResource := &Temurin{
		spec: newSpec,
		apiWebClient: &http.Client{
			Transport: httpClient.Transport,
		},
		apiWebRedirectionClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: httpClient.Transport,
		},
	}

	/** Validations **/
	architectures, err := newResource.apiGetArchitectures()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(architectures, newResource.spec.Architecture) {
		return nil, fmt.Errorf("[temurin] Specified architecture %q is not a valid Temurin architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)", newResource.spec.Architecture)
	}

	operatingSystems, err := newResource.apiGetOperatingSystems()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(operatingSystems, newResource.spec.OperatingSystem) {
		return nil, fmt.Errorf("[temurin] Specified operating system %q is not a valid Temurin architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)", newResource.spec.OperatingSystem)
	}

	if newSpec.ReleaseType != "ea" && newSpec.ReleaseType != "ga" {
		return nil, fmt.Errorf("[temurin] Specified release type %q is invalid: Temurin only accepts 'ga' (stable builds) or 'ea' (nightly builds).", newResource.spec.ReleaseType)
	}

	for _, platform := range newSpec.Platforms {
		splitPlatform := strings.Split(platform, "/")
		if len(splitPlatform) > 2 {
			return nil, fmt.Errorf("[temurin] Specified platform %q is not a valid Temurin platform: too much items specified.", platform)
		}
		if len(splitPlatform) < 2 {
			return nil, fmt.Errorf("[temurin] Specified platform %q is not a valid Temurin platform: it misses a CPU architecture.", platform)
		}
		if !slices.Contains(operatingSystems, splitPlatform[0]) {
			return nil, fmt.Errorf("[temurin] Specified platform %q is not a valid Temurin platform: %q is not a valid Operating System (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list).", platform, splitPlatform[0])
		}
		if !slices.Contains(architectures, splitPlatform[1]) {
			return nil, fmt.Errorf("[temurin] Specified platform %q is not a valid Temurin platform: %q is not a valid Architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list).", platform, splitPlatform[1])
		}

	}

	/** End of validations **/

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (t *Temurin) Changelog() string {
	return fmt.Sprintf("https://adoptium.net/temurin/release-notes/?version=%s\n", t.foundVersion)
}
