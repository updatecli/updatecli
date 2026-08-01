package github

import "regexp"

var githubPullRequestURL = regexp.MustCompile(
	`https://github\.com/([^/\s]+)/([^/\s]+)/pull/(\d+)`,
)

func redirectGitHubPullRequestLinks(body string) string {
	return githubPullRequestURL.ReplaceAllString(
		body,
		"https://redirect.github.com/$1/$2/pull/$3",
	)
}
