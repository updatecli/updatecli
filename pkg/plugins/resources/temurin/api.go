package temurin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

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

func (t Temurin) apiPerformHttpReq(endpoint string, webClient httpclient.HTTPClient) (body []byte, locationHeader string, err error) {
	url := temurinApiUrl + endpoint

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []byte{}, "", fmt.Errorf("something went wrong while performing a request to %q:\n%s\n", redact.URL(url), err)
	}

	req.Header.Set("User-Agent", httputils.UserAgent)

	logrus.Debugf("[temurin] Performing an http GET request to %q...", redact.URL(url))

	res, err := webClient.Do(req)
	if err != nil {
		return []byte{}, "", fmt.Errorf("something went wrong while performing a request to %q:\n%s\n", redact.URL(url), err)
	}
	defer res.Body.Close()

	logrus.Debugf("[temurin] API client returned the following response:\n%v\n", res)

	if res.StatusCode >= 400 {
		_, _ = httputil.DumpResponse(res, false)
		return []byte{}, "", fmt.Errorf("Got an HTTP error %d from the API.\n", res.StatusCode)
	}

	locationHeader = res.Header.Get("Location")
	logrus.Debugf("[temurin] API client got the following Location header value: %q.", locationHeader)

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, "", fmt.Errorf("something went wrong while decoding the answer of the request %q:\n%s\n", redact.URL(url), err)
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

func (t Temurin) apiGetLastFeatureRelease() (result int, err error) {
	apiInfoReleases, err := t.apiGetInfoReleases()
	if err != nil {
		return result, err
	}

	result = apiInfoReleases.MostRecentLTS
	if t.spec.ReleaseLine == "feature" {
		result = apiInfoReleases.MostRecentFeatureRelease
	}

	return result, err
}

func (t Temurin) apiGetInfoReleases() (result *apiInfoReleases, err error) {
	body, err := t.apiGetBody(availableReleasesEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API releases information %q\n", err)
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logrus.Errorf("something went wrong while decoding response from Temurin API releases information%q\n", err)
		return result, err
	}

	return result, err
}

func (t Temurin) apiGetArchitectures() (result []string, err error) {
	body, err := t.apiGetBody(architecturesEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API available architectures %q\n", err)
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logrus.Errorf("something went wrong while decoding response from Temurin API available architectures %q\n", err)
		return result, err
	}

	return result, nil
}

func (t Temurin) apiGetOperatingSystems() (result []string, err error) {
	body, err := t.apiGetBody(osEndpoints)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API available operating systems %q\n", err)
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logrus.Errorf("something went wrong while decoding response from Temurin API operating systems %q\n", err)
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

func (t Temurin) apiGetReleaseNames() (result []string, err error) {
	var versionRange string

	// If user specified a custom version, we have to normalize and validate it
	if t.spec.SpecificVersion != "" {
		parsedVersion, err := t.apiParseVersion(t.spec.SpecificVersion)
		if err != nil {
			return []string{}, err
		}

		versionRange = fmt.Sprintf("(%d.%d.%d, %d.%d.%d]",
			parsedVersion.Major,
			parsedVersion.Minor,
			parsedVersion.Security,
			parsedVersion.Major,
			parsedVersion.Minor,
			parsedVersion.Security+1,
		)

	} else {
		featureVersion := t.spec.FeatureVersion
		if featureVersion == 0 {
			featureVersion, err = t.apiGetLastFeatureRelease()
			if err != nil {
				return []string{}, err
			}
		}

		versionRange = fmt.Sprintf("(%d.0.0, %d.0.0]", featureVersion, featureVersion+1)
	}

	apiEndpoint := fmt.Sprintf(
		"%s?heap_size=normal&image_type=%s&page=0&page_size=10&project=%s&release_type=%s&architecture=%s&os=%s&semver=true&sort_method=DEFAULT&sort_order=DESC&vendor=eclipse&version=%s",
		releaseNamesEndpoint,
		t.spec.ImageType,
		t.spec.Project,
		t.spec.ReleaseType,
		t.spec.Architecture,
		t.spec.OperatingSystem,
		// Mandatory URL encoding otherwise empty responses or HTTP errors
		url.QueryEscape(versionRange),
	)

	logrus.Debugf("[temurin] using API endpoint %q", apiEndpoint)

	body, err := t.apiGetBody(apiEndpoint)
	if err != nil {
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q\n", err)
		return result, err
	}

	var apiResult releaseInformation
	err = json.Unmarshal(body, &apiResult)
	if err != nil {
		logrus.Debugf("[temurin] Failed decoding the response: %q\n", err)
		return result, fmt.Errorf("[temurin] No release found matching provided criteria. Use '--debug' to get details.")
	}

	// Return only the most recent, e.g. the first one (sort is DESC in the URL)
	return apiResult.Releases, nil
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
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q\n", err)
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
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q\n", err)
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
		logrus.Errorf("something went wrong while getting Temurin API latest release information %q\n", err)
		return result, err
	}

	return signatureUrl, nil
}
