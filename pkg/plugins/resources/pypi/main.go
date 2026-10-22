package pypi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"pypi" defines the specification for retrieving PyPI package versions.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "name" defines the PyPI package name.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "name" is required.
	//
	// example:
	//   * name: requests
	//
	Name string `yaml:",omitempty"`
	// "version" defines the package version to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	Version string `yaml:",omitempty"`
	// "url" defines the url of a PyPI compatible registry.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   https://pypi.org/
	//
	URL string `yaml:",omitempty"`
	// "token" defines the bearer token used to authenticate with a private registry.
	//
	// compatible:
	//   * source
	//   * condition
	//
	Token string `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its kind, such as "pep440", "semver" or "latest".
	//
	// compatible:
	//   * source
	//
	// default:
	//   latest
	//
	// remark:
	//   * yanked releases are ignored.
	//   * with the kind "latest", the version is the latest release reported by the registry.
	//     The source fails when that release is yanked.
	//   * with a kind other than "pep440", versions are normalised to semver and dev releases are ignored.
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a release to be considered valid.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * it accepts a duration string, such as "24h", "7d", "3w" or "1y".
	//   * a release's age is based on the earliest upload time of its files.
	//   * releases for which the registry reports no upload time are ignored.
	//   * the lookup fails when the registry reports no upload time for any release.
	//   * the source is skipped when the releases it could return are still too recent,
	//     as they will match later. Releases older than "maximum" never will, so they don't skip it.
	//   * with the kind "latest", the highest release matching the age filter is used
	//     when the latest release reported by the registry doesn't match it.
	//
	// example:
	//   age:
	//     minimum: 7d
	//
	Age age.Spec `yaml:",omitempty"`
}

// releaseFile represents a single distribution file for a release.
type releaseFile struct {
	Yanked     bool   `json:"yanked"`
	UploadTime string `json:"upload_time_iso_8601"`
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
	spec          Spec
	versionFilter version.Filter
	foundVersion  version.Version
	data          pypiData
	webClient     httpclient.HTTPClient
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

	if err := newSpec.Age.Validate(); err != nil {
		return &Pypi{}, fmt.Errorf("wrong age spec: %w", err)
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
func (p *Pypi) getPackageData(ctx context.Context) (pypiData, error) {
	requestURL := fmt.Sprintf("%spypi/%s/json", p.spec.URL, url.PathEscape(p.spec.Name))

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return pypiData{}, fmt.Errorf("building request for %q: %w", requestURL, err)
	}

	if p.spec.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.spec.Token))
	}

	res, err := p.webClient.Do(req)
	if err != nil {
		log.Errorf("fetching pypi package data for %q: %s", p.spec.Name, err)
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

// availableVersions fetches the non-yanked releases of the package, along with how each one
// relates to the age filter. Dev releases are left out unless the versionfilter kind is pep440,
// since they can't be normalized to semver.
func (p *Pypi) availableVersions(ctx context.Context) (releases, error) {
	var err error
	p.data, err = p.getPackageData(ctx)
	if err != nil {
		return nil, err
	}

	rels := releases{}
	for ver, files := range p.data.Releases {
		if isYanked(files) {
			continue
		}

		candidate := ver
		if p.versionFilter.Kind != version.PEP440VERSIONKIND {
			candidate = normalizePEP440(ver)
			if candidate == "" {
				continue // skip dev releases
			}
		}

		status := ageMatching
		if !p.spec.Age.IsZero() {
			status = ageStatus(ver, files, p.spec.Age)
		}
		rels[ver] = release{version: ver, candidate: candidate, status: status}
	}

	/*
		A registry that reports no upload time at all leaves the age filter nothing to work
		with. That isn't a cooldown, it would never end, so it fails rather than skipping
		the source forever without an obvious reason.
	*/
	if !p.spec.Age.IsZero() && len(rels) > 0 && len(rels.withStatus(ageUndated)) == len(rels) {
		return nil, fmt.Errorf("the registry %q doesn't report any upload time for the PyPI package %q, so the age filter can't be applied: remove it for this registry",
			redact.URL(p.spec.URL), p.spec.Name)
	}

	return rels, nil
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
//	1.0.post1  → 1.0.0   (post suffix stripped — SemVer has no equivalent, treated as the base release)
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
	// .post releases have no SemVer equivalent; treat them as the base release.
	if m[2] == "post" {
		return base
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

// getVersions returns the version matching the filter.
func (p *Pypi) getVersions(ctx context.Context) (string, error) {
	rels, err := p.availableVersions(ctx)
	if err != nil {
		return "", err
	}

	versions := rels.candidates(ageMatching)
	coolingDown := rels.candidates(ageTooRecent)

	/*
		The package does publish versions but the age filter discarded every one of them,
		which reports a cooldown still running rather than a lookup failure, so callers are
		expected to skip rather than to fail.
	*/
	if len(versions) == 0 && len(coolingDown) > 0 {
		return "", fmt.Errorf("%w for the PyPI package %q", age.ErrNoVersionMatchingAge, p.spec.Name)
	}

	if p.versionFilter.Kind == version.LATESTVERSIONKIND {
		return p.latestVersion(rels)
	}

	p.foundVersion, err = p.versionFilter.Search(versions)
	if err != nil {
		/*
			Only the versions matching the version filter may be cooling down, such as a
			release published yesterday being the only one satisfying ">=3.0". That is still
			a cooldown, not a lookup failure, so callers are expected to skip rather than to fail.
		*/
		if len(coolingDown) > 0 {
			if _, errAll := p.versionFilter.Search(slices.Concat(versions, coolingDown)); errAll == nil {
				return "", fmt.Errorf("%w for the PyPI package %q", age.ErrNoVersionMatchingAge, p.spec.Name)
			}
		}
		return "", err
	}

	return rels.original(p.foundVersion.GetVersion()), nil
}

// latestVersion returns the latest release reported by the registry or, when the age filter
// keeps it out for now, the highest older release matching the age filter.
func (p *Pypi) latestVersion(rels releases) (string, error) {
	latest := p.data.Info.Version

	rel, published := rels[latest]
	status := rel.status
	switch {
	case !published:
		return "", fmt.Errorf("latest version %q of package %q is yanked", latest, p.spec.Name)
	case status == ageMatching:
		return latest, nil
	case status == ageTooOld:
		return "", fmt.Errorf("latest version %q of package %q is older than the age filter allows", latest, p.spec.Name)
	}

	// The latest release is too recent, or has no upload time to check: an older one may do.
	if found := latestVersionMatchingAge(rels.versions(ageMatching), latest); found != "" {
		log.Debugf("latest version %q of PyPI package %q doesn't match the age filter, using %q", latest, p.spec.Name, found)
		return found, nil
	}

	// A release too recent matches later, which is a cooldown. A missing upload time never appears.
	if status == ageUndated {
		return "", fmt.Errorf("the registry reports no upload time for the latest version %q of package %q, and no older release matches the age filter", latest, p.spec.Name)
	}
	return "", fmt.Errorf("%w for the PyPI package %q", age.ErrNoVersionMatchingAge, p.spec.Name)
}

// ReportConfig returns a sanitized copy of the spec for reporting.
func (p *Pypi) ReportConfig() interface{} {
	return Spec{
		Name:          p.spec.Name,
		Version:       p.spec.Version,
		URL:           redact.URL(p.spec.URL),
		VersionFilter: p.spec.VersionFilter,
		Age:           p.spec.Age,
	}
}
