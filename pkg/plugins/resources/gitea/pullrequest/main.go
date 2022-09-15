package pullrequest

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"

	giteascm "github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
)

// Spec defines settings used to interact with Gitea pullrequest
type Spec struct {
	client.Spec
	// SourceBranch specifies the pullrequest source branch
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// Title defines the Gitea pullrequest title.
	Title string `yaml:",inline,omitempty"`
	// Body defines the Gitea pullrequest body
	Body string `yaml:",inline,omitempty"`
}

// Gitea contains information to interact with Gitea api
type Gitea struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client client.Client
	ge     *giteascm.Gitea
}

// New returns a new valid Gitea object.
func New(spec interface{}, ge *giteascm.Gitea) (Gitea, error) {

	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return Gitea{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return Gitea{}, nil
	}

	if ge != nil {

		if len(clientSpec.Token) == 0 && len(ge.Spec.Token) > 0 {
			clientSpec.Token = ge.Spec.Token
		}

		if len(clientSpec.URL) == 0 && len(ge.Spec.URL) > 0 {
			clientSpec.URL = ge.Spec.URL
		}

		if len(clientSpec.Username) == 0 && len(ge.Spec.Username) > 0 {
			clientSpec.Username = ge.Spec.Username
		}
	}

	// Sanitize modifies the clientSpec so must done one once initialization is finished
	err = clientSpec.Sanitize()
	if err != nil {
		return Gitea{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return Gitea{}, err
	}

	return Gitea{
		spec:   s,
		client: c,
		ge:     ge,
	}, nil

}

func (g *Gitea) CreatePullRequest(title, changelog, pipelineReport string) error {

	var body string
	var sourceBranch string
	var targetBranch string
	var owner string
	var repository string

	body = changelog + "\n" + pipelineReport

	if g.ge != nil {
		sourceBranch = g.ge.HeadBranch
		targetBranch = g.ge.Spec.Branch
		owner = g.ge.Spec.Owner
		repository = g.ge.Spec.Repository
	}

	if len(g.spec.Body) > 0 {
		body = g.spec.Body
	}

	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	if len(g.spec.SourceBranch) > 0 {
		sourceBranch = g.spec.SourceBranch
	}

	if len(g.spec.TargetBranch) > 0 {
		targetBranch = g.spec.TargetBranch
	}

	if len(g.spec.Owner) > 0 {
		owner = g.spec.Owner
	}

	if len(g.spec.Repository) > 0 {
		repository = g.spec.Repository
	}

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

	pullrequests, resp, err := g.client.PullRequests.List(
		ctx,
		strings.Join([]string{
			owner,
			repository}, "/"),
		optsSearch,
	)

	if err != nil {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		return err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	for _, p := range pullrequests {
		if p.Source == sourceBranch &&
			p.Target == targetBranch &&
			!p.Closed &&
			!p.Merged {

			logrus.Infof("%s Nothing else to do, A pullrequest is already opened at:\n\t%s",
				result.SUCCESS,
				p.Link)

			return nil
		}
	}

	// Test that both sourceBranch and targetBranch exists on remote

	ok, err := g.isRemoteBranchesExist()

	if err != nil {
		return err
	}

	/*
		Due to the following scenario, Updatecli always tries to open a Pullrequest
			* A pullrequest has been "manually" closed via UI
			* A previous Updatecli run failed during a Pullrequest creation for example due to network issues


		Therefore we always try to open a PR, we don't consider being an error if all conditions are not met
		such as missing remote branches.
	*/
	if !ok {
		return nil
	}

	opts := scm.PullRequestInput{
		Title:  title,
		Body:   body,
		Source: sourceBranch,
		Target: targetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		sourceBranch,
		targetBranch)

	// Timeout api query after 30sec
	ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := g.client.PullRequests.Create(
		ctx,
		strings.Join([]string{
			owner,
			repository}, "/"),
		&opts,
	)

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		if err.Error() == scm.ErrNotFound.Error() {
			logrus.Infof("Gitea pullrequest not created, skipping")
			return nil
		}
		return err
	}

	logrus.Infof("Gitea pullrequest successfully opened on %q", pr.Link)

	return nil
}

func (g *Gitea) isRemoteBranchesExist() (bool, error) {

	var sourceBranch string
	var targetBranch string
	var owner string
	var repository string

	if g.ge != nil {
		sourceBranch = g.ge.HeadBranch
		targetBranch = g.ge.Spec.Branch
		owner = g.ge.Spec.Owner
		repository = g.ge.Spec.Repository
	}

	if len(g.spec.SourceBranch) > 0 {
		sourceBranch = g.spec.SourceBranch
	}

	if len(g.spec.TargetBranch) > 0 {
		targetBranch = g.spec.TargetBranch
	}

	if len(g.spec.Owner) > 0 {
		owner = g.spec.Owner
	}

	if len(g.spec.Repository) > 0 {
		repository = g.spec.Repository
	}

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	logrus.Printf("%q:%q", g.spec.Owner, g.spec.Repository)

	remoteBranches, resp, err := g.client.Git.ListBranches(
		ctx,
		strings.Join([]string{owner, repository}, "/"),
		scm.ListOptions{
			URL:  g.spec.URL,
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
