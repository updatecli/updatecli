package dockerimage

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/registry"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/changelog/markdown"
)

// Changelog returns the changelog for this resource, or an empty string if not supported
func (di *DockerImage) Changelog(from, to string) *result.Changelogs {

	ref, err := di.createRef(di.foundVersion.GetVersion())
	if err != nil {
		logrus.Debugf("invalid reference %s: %v", di.spec.Image, err)
	}

	manifestData, err := registry.FetchManifest(
		ref.Name(),
		false)
	if err != nil {
		logrus.Debugf("unable to get image %s: %v", di.spec.Image, err)
		return nil
	}

	changelog := getChangelogAnnotation(manifestData)
	if changelog == "" {
		logrus.Debugf("no changelog annotation found in image %s", di.spec.Image)
		return nil
	}

	changelogURL, err := url.Parse(changelog)
	if err == nil && changelogURL.Scheme == "" {
		return nil
	}

	if err != nil {
		logrus.Debugf("unable to parse changelog URL: %v", err)
		return nil
	}

	if !strings.HasSuffix(changelogURL.Path, ".md") {
		logrus.Debugln("As of today changelog must be a markdown available on HTTP/HTTPS")
		return nil
	}

	// Trying to be smart and redirect github url to raw content
	// so we can try to parse it
	// for example "https://github.com/updatecli/policies/tree/main/updatecli/policies/updatecli/autodiscovery/CHANGELOG.md"
	// show be replaced by "
	// https://github.com/updatecli/policies/blob/main/updatecli/policies/updatecli/autodiscovery/CHANGELOG.md?raw=true
	if changelogURL.Host == "github.com" {
		redirectToGitHubRawContent(changelogURL)
	}

	resp, err := http.Get(changelogURL.String())
	if err != nil {
		logrus.Debugf("retrieving changelog from url: %v", err)
		return nil
	}

	defer resp.Body.Close()

	buf := new(strings.Builder)
	// Copy data from the response to standard output
	_, err = io.Copy(buf, resp.Body) //use package "io" and "os"
	if err != nil {
		logrus.Debugf("%v", err)
		return nil
	}

	changelog = buf.String()

	sections, err := markdown.ParseMarkdown([]byte(changelog))
	if err != nil {
		logrus.Debugf("unable to parse changelog: %v", err)
		return nil
	}

	title := di.foundVersion.GetVersion()
	body := sections.GetSectionAsMarkdown(di.foundVersion.GetVersion())

	if body == "" {
		logrus.Debugf("no changelog found for image %s", di.spec.Image)
		return nil
	}

	return &result.Changelogs{
		{
			Title: title,
			Body:  body,
		},
	}

}

// getChangeLogAnnotation returns the changelog annotation from a v1.Descriptor
func getChangelogAnnotation(desc v1.Descriptor) string {

	if changelog, ok := desc.Annotations["org.opencontainers.image.changelog"]; ok {
		return changelog
	}

	return ""
}

// redirectToGitHubRawContent tries to redirect a github url to its associated file raw content
func redirectToGitHubRawContent(u *url.URL) {
	beforePath := u.Path
	if strings.Split(u.Path, "/")[3] == "tree" {
		s := strings.Split(u.Path, "/")
		s[3] = "blob"
		u.Path = strings.Join(s, "/")
	}

	if u.Query().Get("raw") == "" {
		query := u.Query()
		query.Set("raw", "true")
		u.RawQuery = query.Encode()
	}

	if beforePath != u.Path {
		logrus.Debugf("Redirecting %s to %s", beforePath, u.Path)
	}
}
