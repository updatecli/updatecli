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
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	pypiDefaultURL = "https://pypi.org/"
	// legacyUploadTimeLayout is the layout of the "upload_time" field, a UTC time without zone.
	legacyUploadTimeLayout = "2006-01-02T15:04:05"
)

var (
	// pep440PreRelease matches PEP 440 pre-release suffixes: a0, b1, rc2, .post1, .dev3
	pep440PreRelease = regexp.MustCompile(`^(\d+(?:\.\d+)*)(?:\.(post|dev)(\d+)|(a|b|rc)(\d+))$`)
	// pep440ShortRelease matches a final release with fewer than three parts, such as "1" or "1.0".
	pep440ShortRelease = regexp.MustCompile(`^\d+(?:\.\d+)?$`)
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
	//   * the source is skipped, not failed, when the age filter discards every release the
	//     versionfilter matches, too recent or older than "maximum". A versionfilter matching
	//     no release at all still fails.
	//   * the source is skipped as well when the registry reports no upload time for any release.
	//   * with the kind "latest", the highest release matching the age filter is used
	//     when the latest release reported by the registry doesn't match it, which may be
	//     lower than the version in use. The kind "latest" ignores the pattern, so to never
	//     go below that version use the kind "semver" or "pep440" with a pattern bounded at
	//     it, such as ">=2.31".
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
	// LegacyUploadTime is the older "upload_time" field, a UTC time without zone, used when a
	// registry doesn't report "upload_time_iso_8601".
	LegacyUploadTime string `json:"upload_time"`
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
	now := time.Now()
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
			if p.versionFilter.Kind == version.LATESTVERSIONKIND ||
				(p.versionFilter.Kind == version.SEMVERVERSIONKIND && !p.versionFilter.Strict) {
				candidate = padShortRelease(candidate)
			}
		}

		status := ageMatching
		if !p.spec.Age.IsZero() {
			status = ageStatus(ver, files, p.spec.Age, now)
		}
		rels[ver] = newRelease(ver, candidate, status)
	}

	return rels, nil
}

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

// padShortRelease pads a final release with fewer than three parts, such as "1.0" to "1.0.0",
// so that it shares its candidate with its post releases, such as "1.0.post1". It only applies
// to the kinds comparing versions semantically, "semver" and "latest": a "regex" or "time"
// pattern still sees the version as published, and a strict "semver" filter still rejects it.
func padShortRelease(v string) string {
	if !pep440ShortRelease.MatchString(v) {
		return v
	}
	return v + strings.Repeat(".0", 2-strings.Count(v, "."))
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
//
// Like the other resources supporting age, the version filter is applied first and the age
// filter second: when the age filter discards every version the version filter matches, such
// as ">=2.31" while 2.31.0 is still cooling down, it returns age.ErrNoVersionMatchingAge so
// that the source is skipped rather than failed.
func (p *Pypi) getVersions(ctx context.Context) (string, error) {
	rels, err := p.availableVersions(ctx)
	if err != nil {
		return "", err
	}

	// A yanked latest release fails the source, even when the age filter has nothing to work with.
	if p.versionFilter.Kind == version.LATESTVERSIONKIND {
		if err := p.checkLatestPublished(rels); err != nil {
			return "", err
		}
	}

	/*
		A registry reporting no upload time for any release leaves the age filter nothing to
		work with. Like npm, go/module and maven, the source is skipped rather than failed, so
		that an age filter set by default, such as by an autodiscovery crawler, doesn't break
		pipelines that worked without it.
	*/
	if !p.spec.Age.IsZero() && len(rels) > 0 && rels.count(ageUndated) == len(rels) {
		log.Warningf("the registry %q reports no upload time for the PyPI package %q, so the age filter can't be applied: remove it for this registry",
			redact.URL(p.spec.URL), p.spec.Name)
		return "", fmt.Errorf("%w for the PyPI package %q: the registry reports no upload time", age.ErrNoVersionMatchingAge, p.spec.Name)
	}

	if p.versionFilter.Kind == version.LATESTVERSIONKIND {
		return p.latestVersion(rels)
	}

	// kept holds the releases inside the age window, dated every release with an upload time.
	kept := rels.candidates(ageMatching)
	dated := rels.candidates(ageMatching, ageTooRecent, ageTooOld)

	p.foundVersion, err = age.Search(&p.versionFilter, dated, kept)
	if err != nil {
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			return "", fmt.Errorf("%w for the PyPI package %q", err, p.spec.Name)
		}

		// While no release passes the age filter, one the version filter matches may only lack
		// an upload time, which the age filter can't check.
		if !p.spec.Age.IsZero() && len(kept) == 0 &&
			age.MatchesUndated(&p.versionFilter, rels.candidates(ageMatching, ageTooRecent, ageTooOld, ageUndated), dated) {
			return "", fmt.Errorf("%w for the PyPI package %q: the registry reports no upload time for the versions matching %q",
				age.ErrNoVersionMatchingAge, p.spec.Name, p.versionFilter.Pattern)
		}

		return "", err
	}

	return rels.original(p.foundVersion.GetVersion()), nil
}

// latestVersion returns the latest release reported by the registry or, when the age filter
// keeps it out for now, the highest older release matching the age filter. When none does, the
// source is skipped, like npm with the kind "latest".
func (p *Pypi) latestVersion(rels releases) (string, error) {
	latest := p.data.Info.Version

	if err := p.checkLatestPublished(rels); err != nil {
		return "", err
	}

	rel := rels[latest]
	if rel.status == ageMatching {
		return latest, nil
	}

	// The age filter rejects the latest release, or it has no upload time to check: an older one may do.
	found, err := latestVersionMatchingAge(rels.withStatus(ageMatching), rel)
	if err != nil {
		return "", err
	}
	if found != "" {
		log.Debugf("latest version %q of PyPI package %q doesn't match the age filter, using %q", latest, p.spec.Name, found)
		return found, nil
	}

	return "", fmt.Errorf("%w for the PyPI package %q", age.ErrNoVersionMatchingAge, p.spec.Name)
}

// checkLatestPublished returns an error when the latest release reported by the registry is
// yanked, or is a dev release the versionfilter kind leaves out.
func (p *Pypi) checkLatestPublished(rels releases) error {
	latest := p.data.Info.Version
	if _, published := rels[latest]; published {
		return nil
	}

	// Dev releases are left out when they can't be normalized for the versionfilter kind.
	if files, found := p.data.Releases[latest]; found && !isYanked(files) {
		return fmt.Errorf("latest version %q of package %q is a development release, which the %q versionfilter kind ignores", latest, p.spec.Name, p.versionFilter.Kind)
	}
	return fmt.Errorf("latest version %q of package %q is yanked", latest, p.spec.Name)
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
