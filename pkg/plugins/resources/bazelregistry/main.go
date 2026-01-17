package bazelregistry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
	updatecliversion "github.com/updatecli/updatecli/pkg/core/version"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Metadata represents the structure of a Bazel Central Registry metadata.json file
type Metadata struct {
	Versions       []string          `json:"versions"`
	YankedVersions map[string]string `json:"yanked_versions,omitempty"`
	Homepage       string            `json:"homepage,omitempty"`
	Maintainers    []Maintainer      `json:"maintainers,omitempty"`
	Repository     []string          `json:"repository,omitempty"`
}

// Maintainer represents a module maintainer
type Maintainer struct {
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	GitHub       string `json:"github,omitempty"`
	GitHubUserID int    `json:"github_user_id,omitempty"`
}

// Bazelregistry stores configuration about the Bazel registry and the module to query
type Bazelregistry struct {
	spec          Spec
	versionFilter version.Filter // Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	foundVersion  version.Version
	webClient     httpclient.HTTPClient
	baseURL       string
}

// New returns a reference to a newly initialized Bazelregistry object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*Bazelregistry, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()
	if err != nil {
		return nil, err
	}

	// Initialize version filter
	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	// Set default URL if not specified
	baseURL := newSpec.URL
	if baseURL == "" {
		baseURL = DefaultRegistryURL
	}

	// Create HTTP client with timeout
	// The retry client handles retries for transient errors
	retryClient := httpclient.NewRetryClient()
	if httpClient, ok := retryClient.(*http.Client); ok {
		// Set timeout to prevent hanging requests
		httpClient.Timeout = 30 * time.Second
	}

	b := Bazelregistry{
		spec:          newSpec,
		versionFilter: newFilter,
		webClient:     retryClient,
		baseURL:       baseURL,
	}

	return &b, nil
}

// Validate tests that the Bazelregistry struct is correctly configured
func (b *Bazelregistry) Validate() error {
	return b.spec.Validate()
}

// Changelog returns the changelog for this resource, or nil if not supported
func (b *Bazelregistry) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (b *Bazelregistry) ReportConfig() interface{} {
	return Spec{
		Module:        b.spec.Module,
		VersionFilter: b.spec.VersionFilter,
		URL:           redact.URL(b.spec.URL),
	}
}

// getUserAgent returns a User-Agent string including Updatecli version
func getUserAgent() string {
	ua := "updatecli-bazelregistry"
	if updatecliversion.Version != "" {
		ua += "/" + updatecliversion.Version
	} else {
		ua += "/dev"
	}
	return ua
}

// fetchModuleMetadata fetches the metadata.json for a given module.
func (b *Bazelregistry) fetchModuleMetadata(module string) (*Metadata, error) {
	// Build URL by replacing {module} placeholder
	url := strings.ReplaceAll(b.baseURL, "{module}", module)

	logrus.Debugf("Fetching metadata for module %q from %q", module, redact.URL(url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set User-Agent with Updatecli version
	req.Header.Set("User-Agent", getUserAgent())

	resp, err := b.webClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("module %q not found in registry (404)", module)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	metadata, err := parseMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}

	return metadata, nil
}

// parseMetadata parses the metadata.json content
func parseMetadata(data []byte) (*Metadata, error) {
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshaling metadata: %w", err)
	}

	if len(metadata.Versions) == 0 {
		return nil, fmt.Errorf("metadata contains no versions")
	}

	return &metadata, nil
}
