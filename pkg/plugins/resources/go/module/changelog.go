package gomodule

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/sirupsen/logrus"
)

// Changelog returns the changelog for a specific golang module, or an empty string if it couldn't find one
func (g *GoModule) Changelog() string {
	if strings.HasPrefix(g.spec.Path, "github.com") {
		return getChangelogFromGitHub(g.spec.Path, g.foundVersion.OriginalVersion)
	}
	return ""
}

func getChangelogFromGitHub(module, version string) string {
	parsedModule := strings.Split(module, "/")

	if len(parsedModule) != 3 {
		return ""
	}

	URL := fmt.Sprintf("https://api.%s/repos/%s/%s/releases/tags/%s",
		parsedModule[0], parsedModule[1], parsedModule[2], version)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		logrus.Debugf("failed to retrieve changelog from GitHub %q\n", err)
		return ""
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Debugf("failed to retrieve changelog from GitHub %q\n", err)
		return ""
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Debugf("failed to retrieve changelog from GitHub %q\n", err)
		logrus.Debugf("\n%v\n", string(body))
		return ""
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data%q\n", err)
		return ""
	}

	type ReleaseInfo struct {
		HtmlURL string `json:"html_url,"`
		Body    string `json:"body,"`
	}

	release := ReleaseInfo{}

	err = json.Unmarshal(data, &release)
	if err != nil {
		logrus.Errorf("error unmarshalling json: %q", err)
		return ""
	}

	return fmt.Sprintf("Changelog retrieved from:\n\t%s\n%s",
		release.HtmlURL, release.Body)
}
