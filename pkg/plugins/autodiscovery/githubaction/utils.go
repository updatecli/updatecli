package githubaction

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
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
func parseActionName(input string) (URL, owner, repository, directory, reference, kind string) {

	// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/using-pre-written-building-blocks-in-your-workflow

	kind = ACTIONKINDDEFAULT
	if strings.HasPrefix(input, "docker://") {
		return strings.TrimPrefix(input, "docker://"), "", "", "", "", ACTIONKINDDOCKER
	}

	if strings.HasPrefix(input, "./") {
		return "", "", "", "", "", ACTIONKINDLOCAL
	}

	parseURL := strings.Split(input, "@")

	switch len(parseURL) {
	case 0:
		return "", "", "", "", "", ""
	case 1:
		reference = ""
	case 2:
		reference = parseURL[1]
		input = parseURL[0]
	}

	u, err := url.Parse(input)
	if err != nil {
		logrus.Debugf("parsing URL: %s", err)
		return "", "", "", "", "", ""
	}

	path := strings.TrimPrefix(u.Path, "/")
	parseURL = strings.Split(path, "/")

	switch len(parseURL) {
	case 0, 1:
		return "", "", "", "", "", ""
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

		// At this time we expect the URL to be a valid github action such as "actions/checkout"
		if len(p) < 2 {
			logrus.Debugf("unexpected action name %q, feel free to open an issue on the Updatecli project", input)
			return "", "", "", "", "", ""
		}

		owner = p[0]
		repository = p[1]

		if len(p) > 2 {
			directory = strings.Join(p[2:], "/")
		}

		return URL, owner, repository, directory, reference, kind
	}
}

// parseActionDigestComment will parse the action comment from the input string.
// and then try to identify if we already tried to derive a digest
func parseActionDigestComment(input string) (digestReference string) {
	trimmed := strings.TrimSpace(input)
	words := strings.Fields(trimmed)
	if len(words) > 0 {
		return words[0]
	}
	return "" // Return an empty string if no words are found
}

func (ga *GitHubAction) getGitProviderKind(URL string) (kind, token string, err error) {
	if URL == "" {
		URL = defaultGitProviderURL
	}

	u, err := url.Parse(URL)

	if err != nil {
		return "", "", fmt.Errorf("parsing URL: %s", err)
	}

	defaultGitHubToken := func() string {
		if os.Getenv("UPDATECLI_GITHUB_TOKEN") != "" {
			logrus.Debugf("environment  variable UPDATECLI_GITHUB_TOKEN detected, using it values as GitHub token")
			return os.Getenv("UPDATECLI_GITHUB_TOKEN")
		} else if os.Getenv("GITHUB_TOKEN") != "" {
			logrus.Debugf("environment  variable GITHUB_TOKEN detected, using it values as GitHub token")
			return os.Getenv("GITHUB_TOKEN")
		}
		logrus.Debugln("no GitHub token defined, please provide a GitHub token via the setting `token` or one of the environment variable UPDATECLI_GITHUB_TOKEN or GITHUB_TOKEN.")
		return ""
	}

	defaultGiteaToken := func() string {
		if os.Getenv("UPDATECLI_GITEA_TOKEN") != "" {
			logrus.Debugf("environment  variable UPDATECLI_GITEA_TOKEN detected, using it values as Gitea token")
			return os.Getenv("UPDATECLI_GITEA_TOKEN")
		} else if os.Getenv("GITEA_TOKEN") != "" {
			logrus.Debugf("environment  variable GITEA_TOKEN detected, using it values as Gitea token")
			return os.Getenv("GITEA_TOKEN")
		}
		logrus.Debugln("no Gitea token defined, please provide a Gitea token via the setting `token` or one of the environment variable UPDATECLI_GITEA_TOKEN or GITEA_TOKEN.")
		return ""
	}

	for hostname, cred := range ga.credentials {
		if u.Hostname() == hostname {
			kind = cred.Kind
			token = cred.Token
			if token == "" {
				switch kind {
				case kindGitHub:
					token = defaultGitHubToken()
				case kindGitea, kindForgejo:
					token = defaultGiteaToken()
				default:
					logrus.Debugf("unknown kind %q for git provider %q specified in parameter, fallback to github", kind, hostname)
				}
			}
			return kind, token, nil
		}
	}

	if ga.credentials == nil {
		ga.credentials = make(map[string]gitProviderToken)
	}

	switch u.Hostname() {
	case "gitea.com", "codeberg.org", "code.forgejo.org":
		kind = kindGitea
		token = defaultGiteaToken()

		ga.credentials[u.Hostname()] = gitProviderToken{
			Kind:  kind,
			Token: token,
		}

	case "github.com":
		kind = kindGitHub
		token = defaultGitHubToken()
		ga.credentials[u.Hostname()] = gitProviderToken{
			Kind:  kind,
			Token: token,
		}
	default:
		logrus.Debugf("unknown git provider %q, fallback to github", u.Hostname())
		kind = kindGitHub
		token = defaultGitHubToken()
	}

	return kind, token, nil
}
