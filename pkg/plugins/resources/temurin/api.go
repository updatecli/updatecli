package temurin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	httputils "github.com/updatecli/updatecli/pkg/plugins/utils/http"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

const availableReleasesEndpoint = "/info/available_releases"
const architecturesEndpoint = "/types/architectures"
const osEndpoints = "/types/operating_systems"
const installersEndpoint = "/binary/version"
const checksumsEndpoint = "/checksum/version"
const signaturesEndpoint = "/signature/version"
const parseVersionEndpoint = "/version"
const releaseNamesEndpoint = "/info/release_names"

func (t Temurin) baseURL() string {
	if t.apiURL != "" {
		return t.apiURL
	}
	return temurinApiUrl
}

func (t Temurin) apiPerformHttpReq(endpoint string, webClient httpclient.HTTPClient) (body []byte, locationHeader string, err error) {
	url := t.baseURL() + endpoint

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []byte{}, "", fmt.Errorf("something went wrong while performing a request to %q:\n%s", redact.URL(url), err)
	}

	req.Header.Set("User-Agent", httputils.UserAgent)

	logrus.Debugf("[temurin] Performing an http GET request to %q...", redact.URL(url))

	res, err := webClient.Do(req)
	if err != nil {
		return []byte{}, "", fmt.Errorf("something went wrong while performing a request to %q:\n%s", redact.URL(url), err)
	}
	defer res.Body.Close()

	logrus.Debugf("[temurin] API client returned the following response:\n%v\n", res)

	if res.StatusCode >= 400 {
		_, _ = httputil.DumpResponse(res, false)
		return []byte{}, "", fmt.Errorf("got an HTTP error %d from the API", res.StatusCode)
	}

	locationHeader = res.Header.Get("Location")
	logrus.Debugf("[temurin] API client got the following Location header value: %q.", locationHeader)

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, "", fmt.Errorf("something went wrong while decoding the answer of the request %q:\n%s", redact.URL(url), err)
	}

	return body, locationHeader, nil
}

func (t Temurin) apiGetBody(endpoint string) (body []byte, err error) {
	body, _, err = t.apiPerformHttpReq(endpoint, t.apiWebClient)
	return body, err
}

func (t Temurin) apiGetRedirectLocation(endpoint string) (redirectLocation string, err error) {
	_, redirectLocation, err = t.apiPerformHttpReq(endpoint, t.apiWebRedirectionClient)
	return redirectLocation, err
}

func (t Temurin) apiGetLastFeatureRelease() (result int, fallbacks []int, err error) {
	infoReleases, err := t.apiGetInfoReleases()
	if err != nil {
		return result, nil, err
	}

	var candidates []int
	if t.spec.ReleaseLine == "feature" {
		result = infoReleases.MostRecentFeatureRelease
		candidates = slices.Clone(infoReleases.AvailableReleases)
	} else {
		result = infoReleases.MostRecentLTS
		candidates = slices.Clone(infoReleases.AvailableLTSReleases)
	}
	slices.Sort(candidates)

	if len(candidates) == 0 {
		logrus.Debugf("[temurin] API returned no available releases for release line %q", t.spec.ReleaseLine)
	}

	// Build fallbacks: all candidates except the primary, highest version first.
	for i := len(candidates) - 1; i >= 0; i-- {
		if candidates[i] != result {
			fallbacks = append(fallbacks, candidates[i])
		}
	}

	return result, fallbacks, nil
}

func (t Temurin) apiGetInfoReleases() (result *apiInfoReleases, err error) {
	body, err := t.apiGetBody(availableReleasesEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API releases information %q", err)
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logrus.Errorf("something went wrong while decoding response from Temurin API releases information%q", err)
		return result, err
	}

	return result, err
}

func (t Temurin) apiGetArchitectures() (result []string, err error) {
	body, err := t.apiGetBody(architecturesEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API available architectures %q", err)
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logrus.Errorf("something went wrong while decoding response from Temurin API available architectures %q", err)
		return result, err
	}

	return result, nil
}

func (t Temurin) apiGetOperatingSystems() (result []string, err error) {
	body, err := t.apiGetBody(osEndpoints)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API available operating systems %q.", err)
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logrus.Errorf("something went wrong while decoding response from Temurin API operating systems %q.", err)
		return result, err
	}

	return result, nil
}

func (t Temurin) apiParseVersion(version string) (result parsedVersion, err error) {
	apiEndpoint := fmt.Sprintf(
		"%s/%s",
		parseVersionEndpoint,
		version,
	)

	body, err := t.apiGetBody(apiEndpoint)
	if err != nil {
		return result, fmt.Errorf("the version %q is not a valid Temurin version.\nAPI response was: %q", version, err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// apiQueryReleaseNamesForRange queries the release_names endpoint for an arbitrary semver range string.
func (t Temurin) apiQueryReleaseNamesForRange(versionRange string) ([]string, error) {
	apiEndpoint := fmt.Sprintf(
		"%s?heap_size=normal&image_type=%s&page=0&page_size=10&project=%s&release_type=%s&architecture=%s&os=%s&semver=true&sort_method=DEFAULT&sort_order=DESC&vendor=eclipse&version=%s",
		releaseNamesEndpoint,
		url.QueryEscape(t.spec.ImageType),
		url.QueryEscape(t.spec.Project),
		url.QueryEscape(t.spec.ReleaseType),
		url.QueryEscape(t.spec.Architecture),
		url.QueryEscape(t.spec.OperatingSystem),
		url.QueryEscape(versionRange),
	)

	logrus.Debugf("[temurin] using API endpoint %q", apiEndpoint)

	body, err := t.apiGetBody(apiEndpoint)
	if err != nil {
		return nil, err
	}

	var apiResult releaseInformation
	if err := json.Unmarshal(body, &apiResult); err != nil {
		logrus.Debugf("[temurin] Failed decoding the response: %q\n", err)
		return nil, fmt.Errorf("[temurin] No release found matching provided criteria. Use '--debug' to get details")
	}

	return apiResult.Releases, nil
}

// apiQueryReleaseNames queries the release_names endpoint for a specific feature version.
// The version range covers all patch releases within that major version: (N.0.0, N+1.0.0].
func (t Temurin) apiQueryReleaseNames(featureVersion int) ([]string, error) {
	versionRange := fmt.Sprintf("(%d.0.0, %d.0.0]", featureVersion, featureVersion+1)
	return t.apiQueryReleaseNamesForRange(versionRange)
}

func (t Temurin) apiGetReleaseNames() (result []string, err error) {
	// If user specified a custom version, normalize and validate it first.
	if t.spec.SpecificVersion != "" {
		parsedVersion, err := t.apiParseVersion(t.spec.SpecificVersion)
		if err != nil {
			return []string{}, err
		}

		versionRange := fmt.Sprintf("(%d.%d.%d, %d.%d.%d]",
			parsedVersion.Major,
			parsedVersion.Minor,
			parsedVersion.Security,
			parsedVersion.Major,
			parsedVersion.Minor,
			parsedVersion.Security+1,
		)

		return t.apiQueryReleaseNamesForRange(versionRange)
	}

	featureVersion := t.spec.FeatureVersion
	var fallbacks []int
	if featureVersion == 0 {
		featureVersion, fallbacks, err = t.apiGetLastFeatureRelease()
		if err != nil {
			return []string{}, err
		}
	}

	// Try the primary feature version, then fall back to older releases if the
	// API returns no results (e.g. a newly announced major with no GA builds yet).
	const maxAttempts = 3
	candidates := append([]int{featureVersion}, fallbacks...)
	if len(candidates) > maxAttempts {
		candidates = candidates[:maxAttempts]
	}

	var firstErr error
	for i, version := range candidates {
		releases, queryErr := t.apiQueryReleaseNames(version)
		if queryErr != nil {
			if firstErr == nil {
				firstErr = queryErr
				logrus.Debugf("[temurin] feature version %d returned an error, trying fallback: %v", version, queryErr)
			}
			logrus.Debugf("[temurin] falling back from feature version %d after error: %v", version, queryErr)
			continue
		}
		if len(releases) == 0 {
			logrus.Debugf("[temurin] no releases found for feature version %d, trying fallback", version)
			continue
		}
		if i > 0 {
			logrus.Debugf("[temurin] using fallback feature version %d", version)
		}
		return releases, nil
	}

	if firstErr != nil {
		return []string{}, firstErr
	}

	logrus.Debugf("[temurin] exhausted all %d candidate feature versions with no matching releases", len(candidates))
	return []string{}, nil
}

func (t Temurin) apiGetInstallerUrl(releaseName string) (result string, err error) {
	apiEndpoint := fmt.Sprintf(
		"%s/%s/%s/%s/%s/hotspot/normal/eclipse?project=%s",
		installersEndpoint,
		releaseName,
		t.spec.OperatingSystem,
		t.spec.Architecture,
		t.spec.ImageType,
		t.spec.Project,
	)

	logrus.Debugf("[temurin] using API endpoint %q", apiEndpoint)
	locationHeader, err := t.apiGetRedirectLocation(apiEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q.", err)
		return result, err
	}

	return locationHeader, nil
}

func (t Temurin) apiGetChecksumUrl(releaseName string) (result string, err error) {
	apiEndpoint := fmt.Sprintf(
		"%s/%s/%s/%s/%s/hotspot/normal/eclipse?project=%s",
		checksumsEndpoint,
		releaseName,
		t.spec.OperatingSystem,
		t.spec.Architecture,
		t.spec.ImageType,
		t.spec.Project,
	)

	logrus.Debugf("[temurin] using API endpoint %q", apiEndpoint)

	installerChecksumUrl, err := t.apiGetRedirectLocation(apiEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q.", err)
		return result, err
	}

	return installerChecksumUrl, nil
}

func (t Temurin) apiGetSignatureUrl(releaseName string) (result string, err error) {
	apiEndpoint := fmt.Sprintf(
		"%s/%s/%s/%s/%s/hotspot/normal/eclipse?project=%s",
		signaturesEndpoint,
		releaseName,
		t.spec.OperatingSystem,
		t.spec.Architecture,
		t.spec.ImageType,
		t.spec.Project,
	)

	logrus.Debugf("[temurin] using API endpoint %q", apiEndpoint)

	signatureUrl, err := t.apiGetRedirectLocation(apiEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q", err)
		return result, err
	}

	return signatureUrl, nil
}
