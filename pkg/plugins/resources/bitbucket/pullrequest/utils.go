package pullrequest

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/bitbucket/client"
)

type pullRequestDetails struct {
	Number      int
	Title       string
	Description string
	Link        string
}

// isPullRequestExist queries a remote Bitbucket Cloud instance to know if a Pull Request already exists.
func (b *Bitbucket) isPullRequestExist() (exists bool, details pullRequestDetails, err error) {
	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancelList := context.WithTimeout(ctx, 30*time.Second)
	defer cancelList()

	optsSearch := scm.PullRequestListOptions{
		Page:   1,
		Size:   30,
		Open:   true,
		Closed: false,
	}

	pullrequests, resp, err := b.client.PullRequests.List(
		ctx,
		strings.Join([]string{
			b.Owner,
			b.Repository,
		}, "/"),
		optsSearch,
	)
	if err != nil {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		return false, pullRequestDetails{}, err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	for _, p := range pullrequests {
		if p.Source == b.SourceBranch &&
			p.Target == b.TargetBranch &&
			!p.Closed &&
			!p.Merged {

			logrus.Infof("%s Bitbucket pullrequest detected at:\n\t%s",
				result.SUCCESS,
				p.Link)

			return true, pullRequestDetails{
				Number:      p.Number,
				Title:       p.Title,
				Description: p.Body,
				Link:        p.Link,
			}, nil
		}
	}
	return false, pullRequestDetails{}, nil
}

// isRemoteBranchesExist queries a remote Bitbucket Cloud to know if both the pull request source branch and the target branch exist.
func (s *Bitbucket) isRemoteBranchesExist() (bool, error) {
	var sourceBranch string
	var targetBranch string
	var owner string
	var repository string

	if s.scm != nil {
		_, sourceBranch, targetBranch = s.scm.GetBranches()
		owner = s.scm.Spec.Owner
		repository = s.scm.Spec.Repository
	}

	if len(s.spec.SourceBranch) > 0 {
		sourceBranch = s.spec.SourceBranch
	}

	if len(s.spec.TargetBranch) > 0 {
		targetBranch = s.spec.TargetBranch
	}

	if len(s.spec.Owner) > 0 {
		owner = s.spec.Owner
	}

	if len(s.spec.Repository) > 0 {
		repository = s.spec.Repository
	}

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	remoteBranches, resp, err := s.client.Git.ListBranches(
		ctx,
		strings.Join([]string{owner, repository}, "/"),
		scm.ListOptions{
			URL:  client.URL(),
			Page: 1,
			Size: 30,
		},
	)
	if err != nil {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		return false, err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	foundRemoteSourceBranch := false
	foundRemoteTargetBranch := false

	for _, remoteBranch := range remoteBranches {
		if remoteBranch.Name == sourceBranch {
			foundRemoteSourceBranch = true
		}
		if remoteBranch.Name == targetBranch {
			foundRemoteTargetBranch = true
		}

		if foundRemoteSourceBranch && foundRemoteTargetBranch {
			return true, nil
		}
	}

	if !foundRemoteSourceBranch {
		logrus.Debugf("Branch %q not found on remote repository %s/%s",
			sourceBranch,
			owner,
			repository)
	}

	if !foundRemoteTargetBranch {
		logrus.Debugf("Branch %q not found on remote repository %s/%s",
			targetBranch,
			owner,
			repository)
	}

	return false, nil
}

// inheritFromScm retrieve missing bitbucket settings from the bitbucket scm object.
func (s *Bitbucket) inheritFromScm() {
	if s.scm != nil {
		_, s.SourceBranch, s.TargetBranch = s.scm.GetBranches()
		s.Owner = s.scm.Spec.Owner
		s.Repository = s.scm.Spec.Repository
	}

	if len(s.spec.SourceBranch) > 0 {
		s.SourceBranch = s.spec.SourceBranch
	}

	if len(s.spec.TargetBranch) > 0 {
		s.TargetBranch = s.spec.TargetBranch
	}

	if len(s.spec.Owner) > 0 {
		s.Owner = s.spec.Owner
	}

	if len(s.spec.Repository) > 0 {
		s.Repository = s.spec.Repository
	}
}
