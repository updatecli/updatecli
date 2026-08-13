package gomodule

import (
	"context"
	"encoding/json"
	"errors"
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

// ErrNoVersionMatchingAge is returned when a Golang module publishes versions but the age
// filter discarded all of them. It reports a cooldown still running, not a lookup failure,
// so callers are expected to skip rather than to fail.
var ErrNoVersionMatchingAge = errors.New("no version matching the age filter")

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

	// Tracks proxies which do publish versions for that module but none matching the age
	// filter, so that a running cooldown isn't reported as a missing module.
	heldBackByAge := false

	for _, proxy := range strings.Split(GOPROXY, ",") {
		proxy = strings.TrimSpace(proxy)
		if !isSupportedGoProxy(proxy) {
			continue
		}

		publishedVersions, matchingVersions, err := getVersionsFromProxy(ctx, g.webClient, proxy, g.Spec.Module, g.Spec.Age)
		if err != nil {
			logrus.Debugf("skipping proxy %q due to %v\n", proxy, err)
			continue
		}

		/*
			The module doesn't publish any version, so the only thing a Go proxy can report
			is the pseudo version of its latest commit.
			as explained on https://go.dev/ref/mod#goproxy-protocol
		*/
		if len(publishedVersions) == 0 {
			if !isLatestVersionFilter(g.versionFilter) {
				logrus.Debugf("no version published for module %q on proxy %q\n", g.Spec.Module, proxy)
				continue
			}

			latestVersion, err := getLatestVersionFromProxy(ctx, g.webClient, proxy, g.Spec.Module)
			if err != nil {
				logrus.Debugf("skipping proxy %q due to %v\n", proxy, err)
				continue
			}

			if latestVersion.Version == "" {
				logrus.Debugf("no version published for module %q on proxy %q\n", g.Spec.Module, proxy)
				continue
			}

			if !isVersionMatchingAge(latestVersion, g.Spec.Age) {
				logrus.Debugf("ignoring version %q from proxy %q because it doesn't match the age filter\n", latestVersion.Version, proxy)
				heldBackByAge = true
				continue
			}

			logrus.Debugf("no version published for module %q on proxy %q, fallback to version %q\n", g.Spec.Module, proxy, latestVersion.Version)

			g.Version = version.Version{
				ParsedVersion:   latestVersion.Version,
				OriginalVersion: latestVersion.Version,
			}

			return latestVersion.Version, []string{latestVersion.Version}, nil
		}

		// The module publishes versions but the age filter discarded every one of them,
		// which means the version we would have returned is still cooling down.
		if len(matchingVersions) == 0 {
			logrus.Debugf("every version published for module %q on proxy %q is filtered out by the age filter\n", g.Spec.Module, proxy)
			heldBackByAge = true
			continue
		}

		versions = versionNames(matchingVersions)

		/*
			A "latest" filter asks for the most recently published version, which the
			lexicographic sort below can't identify since "v1.9.0" sorts after "v1.10.0".
			Release dates are known whenever the age filter ran, so use them to fall back
			on the most recent version that isn't cooling down anymore.
		*/
		if isNewestVersionFilter(g.versionFilter) && !g.Spec.Age.IsZero() {
			newestVersion := newestPublishedVersion(matchingVersions)
			if newestVersion != "" {
				g.Version = version.Version{
					ParsedVersion:   newestVersion,
					OriginalVersion: newestVersion,
				}

				return newestVersion, versions, nil
			}
		}

		sort.Strings(versions)
		g.Version, err = g.versionFilter.Search(versions)
		if err != nil {
			return "", nil, err
		}

		return g.Version.GetVersion(), versions, nil

	}

	if heldBackByAge {
		return "", nil, fmt.Errorf("%w for GO module %q", ErrNoVersionMatchingAge, g.Spec.Module)
	}

	return "", nil, fmt.Errorf("GO module %q not found on proxy %q", g.Spec.Module, GOPROXY)
}

// getFromProxy queries a Go module proxy endpoint and returns its raw response body.
// The endpoint is built by appending elems to the module path, as described on
// https://go.dev/ref/mod#goproxy-protocol
func getFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module string, elems ...string) ([]byte, error) {
	URL, err := url.JoinPath(
		sanitizeGoProxy(proxy),
		append([]string{sanitizeGoModuleNameForProxy(module)}, elems...)...)
	if err != nil {
		return nil, fmt.Errorf("building go module proxy url: %w", err)
	}

	// #nosec G704
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("building go module proxy request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying go module proxy: %w", err)
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		/*
			A proxy answering with an error status isn't necessarily a problem, GOPROXY may
			list several proxies and the next one is expected to serve the module, so the
			details are only reported at debug level.
		*/
		logrus.Debugf("proxy %q returned HTTP %d (%s) for module %q\n", proxy, res.StatusCode, res.Status, module)

		body, err := httputil.DumpResponse(res, false)
		if err != nil {
			logrus.Debugf("failed to dump proxy response for %q: %v\n", proxy, err)
		} else {
			logrus.Debugf("\n%v\n", string(body))
		}

		return nil, fmt.Errorf("GO module %q not found on proxy %q", module, proxy)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading go module proxy response: %w", err)
	}

	return data, nil
}

// getVersionsFromProxy returns the versions of a Golang module published on a proxy.
//
// It returns both every published version and the subset matching the age filter, as
// telling "this module doesn't publish any version" from "every published version is
// still cooling down" requires very different handling from the caller.
// The matching subset carries release dates because the goproxy protocol doesn't
// guarantee any ordering of the published versions.
func getVersionsFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module string, releaseAge age.Spec) (publishedVersions []string, matchingVersions []versionInfo, err error) {
	data, err := getFromProxy(ctx, client, proxy, module, "@v", "list")
	if err != nil {
		return nil, nil, err
	}

	// The response should be a list of version separated by \n
	// as explained on https://go.dev/ref/mod#goproxy-protocol
	dataStr := strings.TrimSpace(string(data))
	if dataStr == "" {
		return nil, nil, nil
	}

	publishedVersions = strings.Split(dataStr, "\n")

	// Without an age filter there is no reason to pay for one request per version.
	if releaseAge.IsZero() {
		matchingVersions = make([]versionInfo, 0, len(publishedVersions))
		for _, v := range publishedVersions {
			matchingVersions = append(matchingVersions, versionInfo{Version: v})
		}

		return publishedVersions, matchingVersions, nil
	}

	for _, v := range publishedVersions {
		vInfo, err := getVersionInfoFromProxy(ctx, client, proxy, module, v)
		if err != nil {
			logrus.Debugf("ignoring version %q from proxy %q due to %v\n", v, proxy, err)
			continue
		}

		if !isVersionMatchingAge(vInfo, releaseAge) {
			continue
		}

		matchingVersions = append(matchingVersions, vInfo)
	}

	return publishedVersions, matchingVersions, nil
}

// getLatestVersionFromProxy returns the latest version of a Golang module from a proxy
func getLatestVersionFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module string) (versionInfo, error) {
	data, err := getFromProxy(ctx, client, proxy, module, "@latest")
	if err != nil {
		return versionInfo{}, err
	}

	vInfo := versionInfo{}
	if err = json.Unmarshal(data, &vInfo); err != nil {
		return versionInfo{}, fmt.Errorf("something went wrong while parsing go module api data %q", err)
	}

	return vInfo, nil
}

// getVersionInfoFromProxy returns the version information of a Golang module from a proxy
func getVersionInfoFromProxy(ctx context.Context, client httpclient.HTTPClient, proxy, module, version string) (versionInfo, error) {
	data, err := getFromProxy(ctx, client, proxy, module, "@v", version+".info")
	if err != nil {
		return versionInfo{}, err
	}

	vInfo := versionInfo{}
	if err = json.Unmarshal(data, &vInfo); err != nil {
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

// isNewestVersionFilter returns true if the version filter asks for the most recently
// published version rather than for a version matching a pattern.
func isNewestVersionFilter(versionfilter version.Filter) bool {
	return versionfilter.Kind == version.LATESTVERSIONKIND &&
		versionfilter.Pattern == version.LATESTVERSIONKIND
}

// isVersionMatchingAge returns true if a version falls inside the age window.
// Versions without a parsable release date are never considered as matching.
func isVersionMatchingAge(v versionInfo, releaseAge age.Spec) bool {
	if releaseAge.IsZero() {
		return true
	}

	releaseDate, err := time.Parse(time.RFC3339, v.Time)
	if err != nil {
		logrus.Debugf("ignoring version %q due to invalid release date %q: %v\n", v.Version, v.Time, err)
		return false
	}

	if releaseAge.Minimum != "" && releaseAge.IsOlderThan(releaseDate, nil) {
		logrus.Debugf("ignoring version %q because its age is below %q (released on %s)\n", v.Version, releaseAge.Minimum, releaseDate)
		return false
	}

	if releaseAge.Maximum != "" && releaseAge.IsNewerThan(releaseDate, nil) {
		logrus.Debugf("ignoring version %q because its age is above %q (released on %s)\n", v.Version, releaseAge.Maximum, releaseDate)
		return false
	}

	return true
}

// newestPublishedVersion returns the most recently published version among the provided
// ones, or an empty string when none of them carries a parsable release date.
func newestPublishedVersion(versions []versionInfo) string {
	newestVersion := ""
	newestReleaseDate := time.Time{}

	for _, v := range versions {
		releaseDate, err := time.Parse(time.RFC3339, v.Time)
		if err != nil {
			continue
		}

		if newestVersion == "" || releaseDate.After(newestReleaseDate) {
			newestVersion = v.Version
			newestReleaseDate = releaseDate
		}
	}

	return newestVersion
}

// versionNames returns the version strings of the provided version information.
func versionNames(versions []versionInfo) []string {
	names := make([]string, 0, len(versions))
	for _, v := range versions {
		names = append(names, v.Version)
	}

	return names
}
