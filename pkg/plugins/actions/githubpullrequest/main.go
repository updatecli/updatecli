package githubpullrequest

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/options"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Run creates (or updates) the pull request on GitHub
func (p *PullRequest) Run(title, changelog, pipelineReport string) error {
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

// RunTarget commits the change associated to the provided target
func (p *PullRequest) RunTarget(t target.Target, o options.Pipeline) error {
	t.Result = result.ATTENTION
	if !o.DryRun {
		if t.Message == "" {
			t.Result = result.FAILURE
			return fmt.Errorf("target has no change message")
		}

		if len(t.Files) == 0 {
			t.Result = result.FAILURE
			logrus.Info("no changed file to commit")
			return nil
		}

		if o.Commit {
			if err := p.scm.Add(t.Files); err != nil {
				t.Result = result.FAILURE
				return err
			}

			if err := p.scm.Commit(t.Message); err != nil {
				t.Result = result.FAILURE
				return err
			}
		}
		if o.Push {
			if err := p.scm.Push(); err != nil {
				t.Result = result.FAILURE
				return err
			}
		}
	}

	return nil
}
