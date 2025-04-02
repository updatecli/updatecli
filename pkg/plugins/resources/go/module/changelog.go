package gomodule

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
	"golang.org/x/net/html"
)

// Changelog returns the changelog for a specific golang module, or an empty string if it couldn't find one
func (g *GoModule) Changelog(from, to string) *result.Changelogs {

	return DetectChangelogSource(g.Spec.Module, from, to, 0)

}

// DetectChangelogSource tries to identify based on the Goland module
// where the changelog is located. At the moment it only supports
// GitHub repositories both direct and from a proxy.
func DetectChangelogSource(module, from, to string, depth int) *result.Changelogs {

	if module == "" {
		return nil
	}

	if strings.HasPrefix(module, "github.com/") {
		return getChangelogFromGitHub(module, from, to)
	}

	if depth == 0 {
		return getChangelogFromProxy(module, from, to)
	}

	return nil

}

func getChangelogFromProxy(module, from, to string) *result.Changelogs {

	if module == "" {
		return nil
	}

	URL, err := url.Parse(module)
	if err != nil {
		logrus.Errorf("something went wrong while parsing module %q\n", err)
		return nil
	}

	if URL.Scheme == "" {
		URL.Scheme = "https"
	}

	query := URL.Query()
	query.Set("go-get", "1")
	URL.RawQuery = query.Encode()

	httpClient := &http.Client{}

	res, err := httpClient.Get(URL.String())
	if err != nil {
		logrus.Errorf("something went wrong while getting go module api data %q\n", err)
		return nil
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		logrus.Errorf("something went wrong while getting go module data %q\n", err)
		return nil
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	if parsedModule := getGitRepositoryURL(string(data)); parsedModule != "" {
		return DetectChangelogSource(parsedModule, from, to, 1)
	}

	return nil
}

// getChangelogFromGitHub retrieves the releases notes from a GitHub repository
func getChangelogFromGitHub(module, from, to string) *result.Changelogs {
	parsedModule := strings.Split(module, "/")

	if len(parsedModule) < 3 {
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      parsedModule[1],
		Repository: parsedModule[2],
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		logrus.Debugf("ignored error, searching releases: %s", err)
	}

	return &releases
}

// getGitRepositoryURL parse a HTML content and return the golang module source git repository URL
// if it exists.
// https://go.dev/ref/mod#vcs-find
// Please note that we didn't use the xml approach as it couldn't parse the html content.
func getGitRepositoryURL(htmlContent string) string {
	var result string

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var parseMeta func(*html.Node)
	parseMeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}

			if name == "go-import" {
				content = strings.ReplaceAll(content, "\n", " ")
				content = strings.TrimSpace(content)
				// We only support git repositories
				urls := strings.Split(content, " git ")

				if len(urls) > 0 {
					result = urls[1]
				}

				return
			}
		}

		// Recursively parse child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseMeta(c)
		}
	}

	parseMeta(doc)

	URL, err := url.Parse(result)
	if err != nil {
		logrus.Debugf("failed parsing module %q: %s", result, err)
		return ""
	}

	// Remove scheme from URL
	URL.Scheme = ""
	result = strings.TrimPrefix(URL.String(), "//")

	return result
}
