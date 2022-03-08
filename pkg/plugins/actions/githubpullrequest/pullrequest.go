package githubpullrequest

/*
	Keeping the pullrequest code under the github package to avoid cyclic dependencies between github <-> pullrequest
	We need to completely refactor the github package to split the different component into specific sub packages

		github/pullrequest
		github/target
		github/scm
		github/source
		github/condition
*/

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

// PULLREQUESTBODY is the pull request template used as pull request description
// Please note that triple backticks are concatenated with the literals, as they cannot be escaped
const PULLREQUESTBODY = `
# {{ .Title }}

{{ if .Introduction }}
{{ .Introduction }}
{{ end }}


## Report

{{ .Report }}

## Changelog

<details><summary>Click to expand</summary>

` + "````\n{{ .Description }}\n````" + `

</details>

## Remark

This pull request was automatically created using [Updatecli](https://www.updatecli.io).

Please report any issues with this tool [here](https://github.com/updatecli/updatecli/issues/)

`

var (
	ErrAutomergeNotAllowOnRepository = errors.New("automerge is not allowed on repository")
	ErrBadMergeMethod                = errors.New("wrong merge method defined, accepting one of 'squash', 'merge', 'rebase', or ''")
)

// Spec is a specific struct containing github pullrequest settings provided by an updatecli configuration
type Spec struct {
	// Specifies if automerge is enabled for the new pullrequest
	AutoMerge bool `yaml:",omitempty"`
	// Specifies pull request title
	Title string `yaml:",omitempty"`
	// Specifies user input description used during pull body creation
	Description string `yaml:",omitempty"`
	// Specifies repository labels used for pull request. !! Labels must already exist on the repository
	Labels []string `yaml:",omitempty"`
	// Specifies if a pull request is set to draft, default false
	Draft bool `yaml:",omitempty"`
	// Specifies if maintainer can modify pullRequest
	MaintainerCannotModify bool `yaml:",omitempty"`
	// Specifies which merge method is used to incorporate the pull request. Accept "merge", "squash", "rebase", or ""
	MergeMethod string `yaml:",omitempty"`
}

type PullRequest struct {
	gh                *github.Github
	Description       string
	Report            string
	Title             string
	spec              Spec
	remotePullRequest github.PullRequestApi
}

// isMergeMethodValid ensure that we specified a valid merge method.
func isMergeMethodValid(method string) (bool, error) {
	if len(method) == 0 ||
		strings.ToUpper(method) == "SQUASH" ||
		strings.ToUpper(method) == "MERGE" ||
		strings.ToUpper(method) == "REBASE" {
		return true, nil
	}
	logrus.Debugf("%s - %s", method, ErrBadMergeMethod)
	return false, ErrBadMergeMethod
}

// Validate ensure that a pullrequest spec contains validate fields
func (s *Spec) Validate() error {

	if _, err := isMergeMethodValid(s.MergeMethod); err != nil {
		return err
	}
	return nil
}

func NewPullRequest(spec Spec, gh *github.Github) (PullRequest, error) {
	err := spec.Validate()

	return PullRequest{
		gh:   gh,
		spec: spec,
	}, err
}

// generatePullRequestBody generates the body pull request based on PULLREQUESTBODY
func (p *PullRequest) generatePullRequestBody() (string, error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODY))

	buffer := new(bytes.Buffer)

	type params struct {
		Introduction string
		Title        string
		Report       string
		Description  string
	}

	err := t.Execute(buffer, params{
		Introduction: p.spec.Description,
		Description:  p.Description,
		Report:       p.Report,
		Title:        p.Title,
	})

	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// updatePullRequest updates an existing pull request.
func (p *PullRequest) updatePullRequest() error {

	bodyPR, err := p.generatePullRequestBody()
	if err != nil {
		return err
	}

	return p.gh.UpdatePullRequest(p.Title, bodyPR, p.spec.Labels, p.remotePullRequest)
}
