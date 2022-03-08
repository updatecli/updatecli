package githubpullrequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

func (p *PullRequest) Create(title, changelog, pipelineReport string) error {

	p.Description = changelog
	p.Report = pipelineReport
	p.Title = title

	// Check if they are changes that need to be published otherwise exit
	matchingBranch, err := p.gh.NativeGitHandler.IsSimilarBranch(
		p.gh.HeadBranch,
		p.gh.Spec.Branch,
		p.gh.GetDirectory())

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

func (p *PullRequest) Run() error {
	return nil
}

func (p *PullRequest) RunTarget(target target.Target) error {
	return nil
}
