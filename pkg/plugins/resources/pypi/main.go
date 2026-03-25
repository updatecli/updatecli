package pypi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a PyPI package parsed from an updatecli manifest.
type Spec struct {
	// Name defines the PyPI package name.
	Name string `yaml:",omitempty"`
	// Version defines a specific package version for condition checks.
	Version string `yaml:",omitempty"`
	// URL defines the PyPI-compatible registry URL (defaults to https://pypi.org/).
	URL string `yaml:",omitempty"`
	// Token defines the Bearer token for private registries.
	Token string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// releaseFile represents a single distribution file for a release.
type releaseFile struct {
	Yanked bool `json:"yanked"`
}

// packageInfo holds the metadata section of a PyPI JSON response.
type packageInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	ProjectURLs map[string]string `json:"project_urls"`
}

// pypiData is the top-level structure returned by the PyPI JSON API.
type pypiData struct {
	Info     packageInfo              `json:"info"`
	Releases map[string][]releaseFile `json:"releases"`
}

// Pypi defines a resource of kind "pypi".
type Pypi struct {
	spec                 Spec
	versionFilter        version.Filter
	foundVersion         version.Version
	data                 pypiData
	webClient            httpclient.HTTPClient
	normalizedToOriginal map[string]string // maps semver-normalized version back to original PEP 440
}

const pypiDefaultURL = "https://pypi.org/"

// New returns a new valid Pypi resource object.
func New(spec interface{}) (*Pypi, error) {
	var newSpec Spec

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Pypi{}, err
	}

	err = newSpec.Validate()
	if err != nil {
		return &Pypi{}, err
	}

	if newSpec.URL == "" {
		newSpec.URL = pypiDefaultURL
	}

	if !strings.HasSuffix(newSpec.URL, "/") {
		newSpec.URL += "/"
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &Pypi{}, err
	}

	return &Pypi{
		spec:          newSpec,
		versionFilter: newFilter,
		webClient:     httpclient.NewRetryClient(),
	}, nil
}

// Validate checks that the Pypi spec contains required fields.
func (s *Spec) Validate() error {
	if len(s.Name) == 0 {
		return errors.New("pypi package name not defined")
	}
	return nil
}

// getPackageData fetches and parses the PyPI JSON API response for the package.
func (p *Pypi) getPackageData() (pypiData, error) {
	url := fmt.Sprintf("%spypi/%s/json", p.spec.URL, url.PathEscape(p.spec.Name))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return pypiData{}, fmt.Errorf("building request for %q: %w", url, err)
	}

	if p.spec.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.spec.Token))
	}

	res, err := p.webClient.Do(req)
	if err != nil {
		logrus.Errorf("fetching pypi package data for %q: %s", p.spec.Name, err)
		return pypiData{}, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return pypiData{}, fmt.Errorf("pypi API returned HTTP %d for package %q", res.StatusCode, p.spec.Name)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return pypiData{}, fmt.Errorf("reading pypi response body: %w", err)
	}

	var data pypiData
	if err = json.Unmarshal(body, &data); err != nil {
		return pypiData{}, fmt.Errorf("unmarshalling pypi response: %w", err)
	}

	return data, nil
}

// availableVersions returns the list of non-yanked release versions for the package.
// When the filter kind is pep440, raw PEP 440 versions are returned as-is.
// Otherwise, versions are normalized to semver-compatible form and the mapping back
// to the original PEP 440 string is preserved in normalizedToOriginal.
func (p *Pypi) availableVersions() ([]string, error) {
	var err error
	p.data, err = p.getPackageData()
	if err != nil {
		return nil, err
	}

	var versions []string
	if p.versionFilter.Kind != version.PEP440VERSIONKIND {
		p.normalizedToOriginal = make(map[string]string)
	}
	for ver, files := range p.data.Releases {
		if isYanked(files) {
			continue
		}

		if p.versionFilter.Kind == version.PEP440VERSIONKIND {
			versions = append(versions, ver)
			continue
		}

		normalized := normalizePEP440(ver)
		if normalized == "" {
			continue // skip dev releases
		}
		p.normalizedToOriginal[normalized] = ver
		versions = append(versions, normalized)
	}

	return versions, nil
}

// pep440PreRelease matches PEP 440 pre-release suffixes: a0, b1, rc2, .post1, .dev3
var pep440PreRelease = regexp.MustCompile(`^(\d+(?:\.\d+)*)(?:\.(post|dev)(\d+)|(a|b|rc)(\d+))$`)

// normalizePEP440 converts a PEP 440 version string to a semver-compatible form.
// Used for backward compatibility when users configure a semver filter kind — PEP 440
// pre-release strings (e.g. "0.51b0") must be normalized before semver can parse them.
//
//	0.51b0     → 0.51.0-beta.0
//	1.0a1      → 1.0.0-alpha.1
//	2.0rc1     → 2.0.0-rc.1
//	1.0.post1  → 1.0.0+post.1   (mapped to build metadata, sorted after release)
//	1.0.dev3   → (skipped, returns "" — dev releases are excluded)
//	2.31.0     → 2.31.0          (already valid semver, returned as-is)
func normalizePEP440(v string) string {
	m := pep440PreRelease.FindStringSubmatch(v)
	if m == nil {
		return v // already valid or will be handled by semver.NewVersion
	}

	base := m[1]
	// Ensure 3-part base (e.g. "0.51" → "0.51.0")
	if strings.Count(base, ".") < 2 {
		base += ".0"
	}

	// .dev releases are excluded
	if m[2] == "dev" {
		return ""
	}
	// .post releases: map to build metadata so they sort after the release
	if m[2] == "post" {
		return base + "+post." + m[3]
	}

	// Pre-release: a → alpha, b → beta, rc → rc
	preTag := m[4]
	preNum := m[5]
	switch preTag {
	case "a":
		return base + "-alpha." + preNum
	case "b":
		return base + "-beta." + preNum
	case "rc":
		return base + "-rc." + preNum
	}

	return v
}

// isYanked returns true when every file in a release is marked yanked,
// or when the release has no files (empty releases are treated as yanked).
func isYanked(files []releaseFile) bool {
	if len(files) == 0 {
		return true
	}
	for _, f := range files {
		if !f.Yanked {
			return false
		}
	}
	return true
}

// originalVersion maps a normalized version back to the original PEP 440 string.
func (p *Pypi) originalVersion(normalized string) string {
	if orig, ok := p.normalizedToOriginal[normalized]; ok {
		return orig
	}
	return normalized
}

// getVersions returns the version matching the filter and all available versions.
func (p *Pypi) getVersions() (string, []string, error) {
	versions, err := p.availableVersions()
	if err != nil {
		return "", nil, err
	}

	if p.versionFilter.Kind == version.LATESTVERSIONKIND {
		return p.data.Info.Version, versions, nil
	}

	p.foundVersion, err = p.versionFilter.Search(versions)
	if err != nil {
		return "", nil, err
	}

	found := p.foundVersion.GetVersion()
	if p.versionFilter.Kind != version.PEP440VERSIONKIND {
		found = p.originalVersion(found)
	}
	return found, versions, nil
}

// ReportConfig returns a sanitized copy of the spec for reporting.
func (p *Pypi) ReportConfig() interface{} {
	return Spec{
		Name:          p.spec.Name,
		Version:       p.spec.Version,
		URL:           redact.URL(p.spec.URL),
		VersionFilter: p.spec.VersionFilter,
	}
}
