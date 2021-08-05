package jenkins

import (
	"github.com/updatecli/updatecli/pkg/plugins/github"
)

var (
	changelogWarning string = `
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
	g := github.Github{
		Owner:      "jenkinsci",
		Repository: "jenkins",
		Token:      j.Github.Token,
		Username:   j.Github.Username,
	}

	if len(g.Token) == 0 || len(g.Username) == 0 {
		return changelogWarning, nil
	}

	return g.Changelog("jenkins-" + name)
}
