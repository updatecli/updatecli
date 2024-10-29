package githubaction

import (
	"io/fs"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
)

// searchWorkflowFiles will look, recursively, for every files containing a GitHub action workflow from a root directory.
func (g *GitHubAction) searchWorkflowFiles(rootDir string, files []string) error {

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logrus.Debugf("something went wrong while walking in %q: %v\n", path, err)
			return err
		}

		for _, foundFile := range files {
			if !info.IsDir() {
				match, err := filepath.Match(foundFile, info.Name())
				if err != nil {
					continue
				}

				// if file doesn't match the pattern, skip it
				if !match {
					continue
				}

				// Ensure our file is in a .github/workflows directory
				workflow := filepath.Dir(path)
				if filepath.Base(workflow) != "workflows" {
					continue
				}

				workflowDirname := filepath.Dir(workflow)

				if !slices.Contains([]string{".github", ".gitea", ".forgejo"}, filepath.Base(workflowDirname)) {
					continue
				}

				g.workflowFiles = append(g.workflowFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	logrus.Debugf("%d GitHub workflow(s) found", len(g.workflowFiles))

	return nil
}

// parseActionName will parse the action name from the input string.
// and then try to identify which part is the owner and which part is the repository.
func parseActionName(input string) (URL, owner, repository, reference string) {

	parseURL := strings.Split(input, "@")

	switch len(parseURL) {
	case 0:
		return "", "", "", ""
	case 1:
		reference = ""
	case 2:
		reference = parseURL[1]
		input = parseURL[0]
	}

	u, err := url.Parse(input)
	if err != nil {
		logrus.Debugf("parsing URL: %s", err)
		return "", "", "", ""
	}

	path := strings.TrimPrefix(u.Path, "/")
	parseURL = strings.Split(path, "/")

	switch len(parseURL) {
	case 0, 1:
		return "", "", "", ""
	default:

		// for some reason, analyzing an URL without a scheme leads to
		// an empty host and the real host is in the path
		// so we need to check if the first part of the path is a domain
		// cfr test case "GitHub url action without scheme"
		if strings.Contains(parseURL[0], ".") {
			URL = parseURL[0]
			parseURL = parseURL[1:]
			path = strings.Join(parseURL, "/")
		}

		p := strings.Split(path, "/")
		if u.Host != "" {
			URL = u.Host
		}

		if u.Scheme != "" {
			URL = u.Scheme + "://" + URL
		}

		owner = p[0]
		repository = p[1]

		return URL, owner, repository, reference
	}
}
