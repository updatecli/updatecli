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

		if !strings.HasSuffix(url.Path, ".md") {
			logrus.Debugln("As of today changelog must be a markdown available on HTTP/HTTPS")
			return ""
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

		return sections.GetSection(di.foundVersion.GetVersion())
	} else {
		logrus.Debugf("No changelog specified on Updatecli policy: %s", di.spec.Image)
	}

	return ""
}
