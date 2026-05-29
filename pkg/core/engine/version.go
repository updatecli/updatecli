package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/version"
)

// versionHTTPEndpoint is the URL to check for the latest version of updatecli
var versionHTTPEndpoint string = "https://www.updatecli.io/changelogs/updatecli/_index.json"

// CheckLatestPublishedVersion check if the currently used version is the latest version
// available
func CheckLatestPublishedVersion() error {
	client := httpclient.NewThrottledRetryClient(1*time.Second, 1)

	type versionData struct {
		Author      string
		PublishedAt string
		Tag         string
	}

	type responseData struct {
		Latest     versionData
		Changelogs []versionData
	}

	resp, err := client.Get(versionHTTPEndpoint)
	if err != nil {
		return fmt.Errorf("unable to check for the latest version of updatecli: %v", err)
	}

	if resp == nil {
		return fmt.Errorf("unable to check for the latest version of updatecli: response is nil")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read the latest version of updatecli: %v", err)
	}
	defer resp.Body.Close()

	var data responseData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("unable to parse the latest version of updatecli: %v", err)
	}

	// Mean that we are using a development version of updatecli, so we can't compare it with the latest version available
	if version.Version == "" {
		return nil
	}

	currentVersion, err := semver.NewVersion(version.Version)
	if err != nil {
		logrus.Warnf("unable to parse the current version of updatecli: %v", err)
		return nil
	}

	latestVersion, err := semver.NewVersion(data.Latest.Tag)
	if err != nil {
		logrus.Warnf("unable to parse the latest version of updatecli: %v", err)
		return nil
	}

	if currentVersion.GreaterThanEqual(latestVersion) {
		return nil
	}

	logrus.Infof("| A new release is available: %q -> %q", currentVersion.String(), latestVersion.String())
	logrus.Infof("| More information on https://www.updatecli.io/changelogs/updatecli/changelogs/%s/", data.Latest.Tag)

	return nil
}
