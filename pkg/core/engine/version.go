package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

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

	sanitizeVersion := func(v string) string {
		s := strings.TrimSpace(v)
		s = strings.TrimPrefix(s, "v")

		return s
	}

	if sanitizeVersion(data.Latest.Tag) == sanitizeVersion(version.Version) {
		return nil
	}

	logrus.Infof("\n---")
	logrus.Infof("A new version of updatecli is available: %s (current: %s)", data.Latest.Tag, version.Version)
	logrus.Infof("Changelog available at: www.updatecli.io/changelogs/updatecli/changelogs/%s/", data.Latest.Tag)
	logrus.Infof("---")

	return nil
}
