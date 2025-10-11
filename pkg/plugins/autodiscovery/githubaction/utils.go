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
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
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
// It will also try to identify if the action is a docker image or a local action
// More information on:
//   - https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsuses
//   - https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/using-pre-written-building-blocks-in-your-workflow#referencing-a-container-on-docker-hub
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

	defaultGitHubToken := func(hostname string) string {

		// 1. Get the token source from the environment
		_, sourceToken, err := github.GetTokenSourceFromEnv()
		if err != nil {
			logrus.Debugf("no GitHub token found in environment variables: %s", err)
		}

		// 2. Get the token from the configuration
		if sourceToken == nil && ga.spec.Credentials[hostname].Token != "" {
			logrus.Debugf("using GitHub token from configuration")
			return ga.spec.Credentials[hostname].Token
		}

		// 3. Get the token from the GitHub App configuration
		if sourceToken == nil && ga.spec.Credentials[hostname].App != nil {
			sourceToken, err = ga.spec.Credentials[hostname].App.Getoauth2TokenSource()
			if err != nil {
				logrus.Debugf("failed to get oauth2 token source from GitHub App spec: %s", err)
			} else {
				logrus.Debugf("using GitHub App authentication from configuration")
			}
		}

		// 4. Fallback to the GITHUB_TOKEN environment variable if no other token source could be found
		if sourceToken == nil {
			_, sourceToken = github.GetFallbackTokenSourceFromEnv()
			if sourceToken != nil {
				logrus.Debugf("using GitHub token from environment variable GITHUB_TOKEN")
			}
		}

		// 5. If we still don't have a token source, log a message and return an empty string
		if sourceToken == nil {
			logrus.Debugln("no GitHub token defined, please provide a GitHub token via the setting `token` or one of the environment variable UPDATECLI_GITHUB_TOKEN or GITHUB_TOKEN.")
			return ""
		}

		accessToken, err := github.GetAccessToken(sourceToken)
		if err != nil {
			logrus.Debugf("failed to get access token from token source: %s", err)
			return ""
		}

		return accessToken
	}

	defaultGiteaToken := func() string {
		if os.Getenv("UPDATECLI_GITEA_TOKEN") != "" {
			logrus.Debugf("environment  variable UPDATECLI_GITEA_TOKEN detected, using it values as Gitea token")
			return os.Getenv("UPDATECLI_GITEA_TOKEN")
		} else if os.Getenv("GITEA_TOKEN") != "" {
			logrus.Debugf("environment  variable GITEA_TOKEN detected, using it values as Gitea token")
			return os.Getenv("GITEA_TOKEN")
		} else if os.Getenv("GITHUB_TOKEN") != "" {
			logrus.Debugf("environment  variable GITHUB_TOKEN detected, using it values as Gitea token")
			return os.Getenv("GITHUB_TOKEN")
		}
		logrus.Debugln("no Gitea token defined, please provide a Gitea token via the setting `token` or one of the environment variable UPDATECLI_GITEA_TOKEN or GITEA_TOKEN.")
		return ""
	}

	for hostname, cred := range ga.credentials {
		if u.Hostname() == hostname {
			kind = cred.Kind
			switch kind {
			case kindGitHub:
				token = defaultGitHubToken(hostname)
			case kindGitea, kindForgejo:
				token = cred.Token
				if token == "" {
					token = defaultGiteaToken()
				}
			default:
				logrus.Debugf("unknown kind %q for git provider %q specified in parameter, fallback to github", kind, hostname)
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
		token = defaultGitHubToken("github.com")
		ga.credentials[u.Hostname()] = gitProviderToken{
			Kind:  kind,
			Token: token,
		}
	default:
		logrus.Debugf("unknown git provider %q, fallback to github", u.Hostname())
		kind = kindGitHub
		token = defaultGitHubToken("github.com")
	}

	return kind, token, nil
}
