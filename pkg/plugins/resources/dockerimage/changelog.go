package dockerimage

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	//"github.com/updatecli/updatecli/pkg/plugins/changelogs/markdown"

	"github.com/updatecli/updatecli/pkg/core/registry"
	"github.com/updatecli/updatecli/pkg/plugins/changelog/markdown"
)

// Changelog returns the changelog for this resource, or an empty string if not supported
func (di *DockerImage) Changelog() string {

	ref, err := di.createRef(di.foundVersion.GetVersion())
	if err != nil {
		logrus.Debugf("invalid reference %s: %v", di.spec.Image, err)
	}

	_, manifestData, err := registry.FetchManifest(ref.Name(), false)
	if err != nil {
		logrus.Debugf("unable to get image %s: %v", di.spec.Image, err)
		return ""
	}

	if changelog, ok := manifestData.Annotations["org.opencontainers.image.changelog"]; ok {

		url, err := url.Parse(changelog)
		if err == nil && url.Scheme == "" {
			return ""
		}

		if err != nil {
			logrus.Debugf("unable to parse changelog URL: %v", err)
			return ""
		}

		if !strings.HasSuffix(url.Path, ".md") {
			logrus.Debugln("As of today changelog must be a markdown available on HTTP/HTTPS")
			return ""
		}

		// Trying to be smart and redirect github url to raw content
		// so we can try to parse it
		// for example "https://github.com/updatecli/policies/tree/main/updatecli/policies/updatecli/autodiscovery/CHANGELOG.md"
		// show be replaced by "
		// https://github.com/updatecli/policies/blob/main/updatecli/policies/updatecli/autodiscovery/CHANGELOG.md?raw=true
		if url.Host == "github.com" {
			beforePath := url.Path
			if strings.Split(url.Path, "/")[3] == "tree" {
				s := strings.Split(url.Path, "/")
				s[3] = "blob"
				url.Path = strings.Join(s, "/")
			}

			if url.Query().Get("raw") == "" {
				url.Query().Add("raw", "true")
			}

			if beforePath != url.Path {
				logrus.Debugf("Redirecting %s to %s", beforePath, url.Path)
			}
		}

		resp, err := http.Get(url.String())
		if err != nil {
			logrus.Debugf("retrieving changelog from url: %v", err)
			return ""
		}

		defer resp.Body.Close()

		buf := new(strings.Builder)
		// Copy data from the response to standard output
		_, err = io.Copy(buf, resp.Body) //use package "io" and "os"
		if err != nil {
			logrus.Debugf("%v", err)
			return ""
		}

		changelog = buf.String()

		sections, err := markdown.ParseMarkdown([]byte(changelog))
		if err != nil {
			logrus.Debugf("unable to parse changelog: %v", err)
			return ""
		}

		return sections.GetSectionAsMarkdown(di.foundVersion.GetVersion())
	}

	return ""
}
