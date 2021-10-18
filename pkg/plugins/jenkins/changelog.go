package jenkins

import (
	"github.com/updatecli/updatecli/pkg/plugins/github"
)

var (
	// ChangelogFallbackContent is returned when no Github credentials are provided as it's currently
	// the only to retrieve Jenkins changelog from GitHub release.
	ChangelogFallbackContent string = `
Warning:
========
We currently rely on Github Release to get Jenkins changelog 
which requires a github token and a github username as in the following example:
---
source:
  kind: jenkins
  spec:
    release: stable
    github:
      token: {{ requiredEnv .github.token }}
      username: {{ .github.username }}
---
`
)

// Changelog returns a changelog description based on a Jenkins version
// retrieved from a GitHub Release
func (j *Jenkins) Changelog(name string) (string, error) {
	if len(j.Github.Token) == 0 || len(j.Github.Username) == 0 {
		return ChangelogFallbackContent, nil
	}

	g, err := github.New(github.Spec{
		Owner:      "jenkinsci",
		Repository: "jenkins",
		Token:      j.Github.Token,
		Username:   j.Github.Username})

	if err != nil {
		return "", err

	}
	changelog, err := g.Changelog("jenkins-" + name)

	return changelog, err
}
