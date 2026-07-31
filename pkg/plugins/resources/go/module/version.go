package gomodule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// versionInfo represents the structure of the version information returned by the Go proxy API.
type versionInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

// GetVersions fetch all versions of a Golang module
func (g *GoModule) versions(ctx context.Context) (v string, versions []string, err error) {

	var GOPROXY string
	if g.Spec.Proxy != "" {
		GOPROXY = g.Spec.Proxy
	} else if os.Getenv("GOPROXY") != "" {
		GOPROXY = os.Getenv("GOPROXY")
	} else {
		GOPROXY = goModuleDefaultProxy
	}

	for _, proxy := range strings.Split(GOPROXY, ",") {
		proxy = strings.TrimSpace(proxy)
		if !isSupportedGoProxy(proxy) {
			continue
		}

		proxyVersions, err := getVersionsFromProxy(ctx, g.webClient, proxy, g.Spec.Module, g.Spec.Age)
		if err != nil {
			logrus.Debugf("skipping proxy %q due to %v\n", proxy, err)
			continue
		}

		if proxyVersions == nil && isLatestVersionFilter(g.versionFilter) {
			pseudoVersion, err := getLatestVersionFromProxy(ctx, g.webClient, proxy, g.Spec.Module)
			if err != nil {
				logrus.Debugf("skipping proxy %q due to %v\n", proxy, err)
				continue
			}

			if pseudoVersion != "" {

				if !isPseudoVersionMatchingAge(pseudoVersion, g.Spec.Age) {
					logrus.Debugf("ignoring pseudo version %q from proxy %q because it doesn't match the age filter\n", pseudoVersion, proxy)
					continue
				}

				versions = append(versions, pseudoVersion)
			}

			logrus.Debugf("no version published for module %q on proxy %q, fallback to pseudo version %q\n", g.Spec.Module, proxy, pseudoVersion)

			return pseudoVersion, versions, nil
		}

		/*
			The response should be a list of version separated by \n
			as explained on https://go.dev/ref/mod#goproxy-protocol
		*/
		versions = append(versions, proxyVersions...)

		sort.Strings(versions)
		g.Version, err = g.versionFilter.Search(versions)
		if err != nil {
			return "", nil, err
		}

		return g.Version.GetVersion(), versions, nil

	}

	return "", nil, fmt.Errorf("GO module %q not found on proxy %q", g.Spec.Module, GOPROXY)
}

// getVersionsFromProxy returns all versions of a Golang module from a proxy
func getVersionsFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module string, releaseAge age.Spec) ([]string, error) {
	URL, err := url.JoinPath(
		sanitizeGoProxy(proxy),
		sanitizeGoModuleNameForProxy(module),
		"@v", "list")
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return nil, err
	}

	// #nosec G704
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		logrus.Errorf("something went wrong while getting golang module data: proxy %q returned HTTP %d (%s)\n", proxy, res.StatusCode, res.Status)
		logrus.Debugf("skipping proxy %q due to HTTP %d (%s)\n", proxy, res.StatusCode, res.Status)
		body, err := httputil.DumpResponse(res, false)
		if err != nil {
			logrus.Debugf("failed to dump proxy response for %q: %q\n", proxy, err)
		} else {
			logrus.Debugf("\n%v\n", string(body))
		}

		return nil, fmt.Errorf("GO module %q not found on proxy %q", module, proxy)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting golang module api data%q\n", err)
		return nil, err
	}

	// The response should be a list of version separated by \n
	// as explained on https://go.dev/ref/mod#goproxy-protocol

	dataStr := strings.TrimSpace(string(data))
	versions := strings.Split(dataStr, "\n")

	// Sanitize versions by filtering out versions that are too recent based on the MinimumReleaseAge filter
	if !releaseAge.IsZero() {
		sanitizedVersions := []string{}

		for v := range versions {
			getVersionInfo, err := getVersionInfoFromProxy(ctx, client, proxy, module, versions[v])
			if err != nil {
				logrus.Debugf("ignoring version %q from proxy %q due to %q\n", versions[v], proxy, err)
				continue
			}

			releaseDate, err := time.Parse(time.RFC3339, getVersionInfo.Time)
			if err != nil {
				logrus.Debugf("ignoring version %q from proxy %q due to invalid release date format: %q\n", versions[v], proxy, err)
				continue
			}

			if releaseAge.Minimum != "" && releaseAge.IsOlderThan(releaseDate, nil) {
				logrus.Debugf("ignoring	version %q from proxy %q because its age is below %q (released on %s)\n", versions[v], proxy, releaseAge.Minimum, releaseDate)
				continue
			}

			if releaseAge.Maximum != "" && releaseAge.IsNewerThan(releaseDate, nil) {
				logrus.Debugf("ignoring version %q from proxy %q because its age is above %q (released on %s)\n", versions[v], proxy, releaseAge.Maximum, releaseDate)
				continue
			}

			sanitizedVersions = append(sanitizedVersions, versions[v])
		}
		versions = sanitizedVersions
	}

	if len(versions) == 0 {
		return nil, nil
	}

	if len(versions) == 1 && versions[0] == "" {
		return nil, nil
	}

	return versions, nil
}

// getLatestVersionFromProxy returns the latest version of a Golang module from a proxy
func getLatestVersionFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module string) (string, error) {
	URL, err := url.JoinPath(
		sanitizeGoProxy(proxy),
		sanitizeGoModuleNameForProxy(module),
		"@latest")

	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return "", err
	}

	// #nosec G704
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Errorf("something went wrong while getting golang module data %q\n", err)
		logrus.Debugf("skipping proxy %q due to %q\n", proxy, err)
		logrus.Debugf("\n%v\n", string(body))

		return "", fmt.Errorf("GO module %q not found on proxy %q", module, proxy)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data%q\n", err)
		return "", err
	}

	type JSONData struct {
		Version string `json:"Version"`
		Time    string `json:"Time"`
	}

	jsonData := JSONData{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return "", fmt.Errorf("something went wrong while parsing go module api data %q", err)
	}

	return jsonData.Version, nil
}

// getVersionInfoFromProxy returns the version information of a Golang module from a proxy
func getVersionInfoFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module, version string) (versionInfo, error) {
	URL, err := url.JoinPath(
		sanitizeGoProxy(proxy),
		sanitizeGoModuleNameForProxy(module),
		"@v",
		version+".info")

	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return versionInfo{}, err
	}

	// #nosec G704
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return versionInfo{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return versionInfo{}, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Errorf("something went wrong while getting golang module data %q\n", err)
		logrus.Debugf("skipping proxy %q due to %q\n", proxy, err)
		logrus.Debugf("\n%v\n", string(body))

		return versionInfo{}, fmt.Errorf("GO module %q not found on proxy %q", module, proxy)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting go module proxy response data%q\n", err)
		return versionInfo{}, err
	}

	vInfo := versionInfo{}
	err = json.Unmarshal(data, &vInfo)
	if err != nil {
		return versionInfo{}, fmt.Errorf("something went wrong while parsing go module api data %q", err)
	}

	return vInfo, nil
}

// isLatestVersionFilter returns true if the version filter is looking for the latest version.
func isLatestVersionFilter(versionfilter version.Filter) bool {

	if versionfilter.Kind == version.LATESTVERSIONKIND {
		return true
	}

	if versionfilter.Kind == version.SEMVERVERSIONKIND && versionfilter.Pattern == "*" {
		return true
	}

	if versionfilter.Kind == version.SEMVERVERSIONKIND && versionfilter.Pattern == "" {
		return true
	}

	if versionfilter.Kind == version.SEMVERVERSIONKIND && versionfilter.Pattern == ">=0.0.0-0" {
		return true
	}

	return false
}

func isPseudoVersionMatchingAge(pseudoVersion string, releaseAge age.Spec) bool {
	// Pseudo versions have the format vX.0.0-yyyymmddhhmmss-abcdefabcdef
	parts := strings.Split(pseudoVersion, "-")
	if len(parts) < 3 {
		logrus.Debugf("invalid pseudo version format: %q\n", pseudoVersion)
		return false
	}

	timestamp := parts[len(parts)-2]
	releaseDate, err := time.Parse("20060102150405", timestamp)
	if err != nil {
		logrus.Debugf("failed to parse release date from pseudo version %q: %q\n", pseudoVersion, err)
		return false
	}

	if releaseAge.Minimum != "" && releaseAge.IsOlderThan(releaseDate, nil) {
		logrus.Debugf("ignoring pseudo version %q because its age is below %q (released on %s)\n", pseudoVersion, releaseAge.Minimum, releaseDate)
		return false
	}

	if releaseAge.Maximum != "" && releaseAge.IsNewerThan(releaseDate, nil) {
		logrus.Debugf("ignoring pseudo version %q because its age is above %q (released on %s)\n", pseudoVersion, releaseAge.Maximum, releaseDate)
		return false
	}

	return true
}
