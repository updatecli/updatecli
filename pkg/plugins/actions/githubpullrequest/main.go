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
	"fmt"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/options"
	"github.com/updatecli/updatecli/pkg/core/result"
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

Please report any issues with this tool [here](https://github.com/updatecli/updatecli/issues/new)

`

var (
	ErrBadMergeMethod = errors.New("wrong merge method defined, accepting one of 'squash', 'merge', 'rebase', or ''")
)

// Spec is a specific struct containing pullrequest settings provided
// by an updatecli configuration
type Spec struct {
	AutoMerge              bool     // Specify if automerge is enabled for the new pullrequest
	Title                  string   // Specify pull request title
	Description            string   // Description contains user input description used during pull body creation
	Labels                 []string // Specify repository labels used for pull request. !! They must already exist
	Draft                  bool     // Define if a pull request is set to draft, default false
	MaintainerCannotModify bool     // Define if maintainer can modify pullRequest
	MergeMethod            string   // Define which merge method is used to incorporate the pull request. Accept "merge", "squash", "rebase", or ""
}

type PullRequest struct {
	gh                *github.Github
	Description       string
	Report            string
	Title             string
	spec              Spec
	remotePullRequest github.PullRequestApi
	options           options.Pipeline
	// scm               scm.ScmHandler
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

func NewPullRequest(spec Spec, gh *github.Github, opts options.Pipeline) (PullRequest, error) {
	err := spec.Validate()

	return PullRequest{
		gh:      gh,
		spec:    spec,
		options: opts,
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

// Run creates (or updates) the pull request on GitHub
func (p *PullRequest) Run(title, changelog, pipelineReport string) error {
	p.Description = changelog
	p.Report = pipelineReport
	p.Title = title

	// When dry run is enabled (e.g. updatecli diff <...>), only prints messages with expected state and returns
	if p.options.DryRun {
		logrus.Infof("[Dry Run] A GitHub Pull Request with the title %q is expected.", p.Title)

		actionDebugOutput := fmt.Sprintf("The expected pull request would have the following informations:\n\n##Title:\n%s\n\n##Changelog:\n\n%s\n\n##Report:\n\n%s\n\n=====\n",
			p.Title,
			p.Description,
			p.Report,
		)
		logrus.Debugf(strings.ReplaceAll(actionDebugOutput, "\n", "\n\t|\t"))

		return nil
	}

	// Start by pushing changes to the remote
	if p.options.Push {
		if err := p.gh.Push(); err != nil {
			return err
		}
	}

	// Check if they are changes that need to be published otherwise exit
	matchingBranch, err := p.gh.NativeGitHandler.IsSimilarBranch(
		p.gh.HeadBranch,
		p.gh.Spec.Branch,
		p.gh.GetDirectory(),
	)
	if err != nil {
		return err
	}

	if matchingBranch {
		logrus.Debugf("No changes detected between branches %q and %q, skipping pullrequest creation",
			p.gh.HeadBranch,
			p.gh.Spec.Branch)

		return nil
	}

	// Check if there is already a pullRequest for current pipeline
	p.remotePullRequest, err = p.gh.GetRemotePullRequest()
	if err != nil {
		return err
	}

	bodyPR, err := p.generatePullRequestBody()
	if err != nil {
		return err
	}

	// No pullrequest ID so we first have to create the remote pullrequest
	if len(p.remotePullRequest.ID) == 0 {
		p.remotePullRequest, err = p.gh.OpenPullRequest(
			p.Title,
			bodyPR,
			p.spec.MaintainerCannotModify,
			p.spec.Draft,
		)
		if err != nil {
			return err
		}
	}

	// Once the remote pull request exist, we can than update it with additional information such as
	// tags,assignee,etc.

	err = p.updatePullRequest()
	if err != nil {
		return err
	}

	if p.spec.AutoMerge {
		err = p.gh.EnablePullRequestAutoMerge(p.spec.MergeMethod, p.remotePullRequest)
		if err != nil {
			return err
		}
	}

	return nil
}

// RunTarget commits the change associated to the provided target
func (p PullRequest) RunTarget(message string, files []string) (string, error) {
	if message == "" {
		return result.FAILURE, fmt.Errorf("target has no change message")
	}

	if len(files) == 0 {
		logrus.Info("no changed file to commit")
		return result.FAILURE, fmt.Errorf("no changed file to commit")
	}

	if p.options.Commit {
		if err := p.gh.Add(files); err != nil {
			return result.FAILURE, err
		}

		if err := p.gh.Commit(message); err != nil {
			return result.FAILURE, err
		}
	}

	return result.ATTENTION, nil
}

func (p PullRequest) InitWorkingDir() (string, error) {
	if err := p.gh.Checkout(); err != nil {
		return "", err
	}

	return p.gh.GetDirectory(), nil
}
