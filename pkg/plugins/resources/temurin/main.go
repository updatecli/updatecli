package temurin

import (
	"fmt"
	"net/http"
	"slices"

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

	newResource := &Temurin{
		spec:         newSpec,
		apiWebClient: &http.Client{},
		apiWebRedirectionClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	/** Validations **/
	architectures, err := newResource.apiGetArchitectures()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(architectures, newResource.spec.Architecture) {
		return nil, fmt.Errorf("[temurin] Specified architecture %q is not a valid Temurin architecture (%v)", newResource.spec.Architecture, architectures)
	}

	operatingSystems, err := newResource.apiGetOperatingSystems()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(operatingSystems, newResource.spec.OperatingSystem) {
		return nil, fmt.Errorf("[temurin] Specified operating system ('os') %q is not a valid Temurin architecture (%v)", newResource.spec.OperatingSystem, operatingSystems)
	}

	if newSpec.ReleaseType != "ea" && newSpec.ReleaseType != "ga" {
		return nil, fmt.Errorf("[temurin] Specified release type %q is invalid: Temurin only accepts 'ga' (stable builds) or 'ea' (nightly builds).", newResource.spec.ReleaseType)
	}

	/** End of validations **/

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (t *Temurin) Changelog() string {
	return fmt.Sprintf("https://adoptium.net/temurin/release-notes/?version=%s\n", t.foundVersion)
}
