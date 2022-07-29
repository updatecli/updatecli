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
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
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
		Spec:   s,
		client: c,
		ge:     ge,
	}, nil

}

func (g *Gitea) CreatePullRequest(title, changelog, pipelineReport string) error {

	var pullrequestBody string
	var pullrequestTitle string
	var pullrequestSourceBranch string
	var pullrequestTargetBranch string
	var pullrequestRepositoryOwner string
	var pullrequestRepositoryName string

	pullrequestBody = changelog + "\n" + pipelineReport
	pullrequestTitle = title

	if g.ge != nil {
		pullrequestSourceBranch = g.ge.HeadBranch
		pullrequestTargetBranch = g.ge.Spec.Branch
		pullrequestRepositoryOwner = g.ge.Spec.Owner
		pullrequestRepositoryName = g.ge.Spec.Repository
	}

	if len(g.Spec.Body) > 0 {
		pullrequestBody = g.Spec.Body
	}

	if len(g.Spec.Title) > 0 {
		pullrequestTitle = g.Spec.Title
	}

	if len(g.Spec.SourceBranch) > 0 {
		pullrequestSourceBranch = g.Spec.SourceBranch
	}

	if len(g.Spec.TargetBranch) > 0 {
		pullrequestTargetBranch = g.Spec.TargetBranch
	}

	if len(g.Spec.Owner) > 0 {
		pullrequestRepositoryOwner = g.Spec.Owner
	}

	if len(g.Spec.Repository) > 0 {
		pullrequestRepositoryName = g.Spec.Repository
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
			pullrequestRepositoryOwner,
			pullrequestRepositoryName}, "/"),
		optsSearch,
	)

	if err != nil {
		return err
	}

	if resp.Status > 400 {
		logrus.Debugf("Gitea Api Response: %v", resp)
	}

	for _, p := range pullrequests {
		if p.Source == pullrequestSourceBranch &&
			p.Target == pullrequestTargetBranch &&
			!p.Closed &&
			!p.Merged {

			logrus.Infof("%s Nothing else to do, A pullrequest is already opened at:\n\t%s",
				result.SUCCESS,
				p.Link)

			return nil
		}
	}

	opts := scm.PullRequestInput{
		Title:  pullrequestTitle,
		Body:   pullrequestBody,
		Source: pullrequestSourceBranch,
		Target: pullrequestTargetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		pullrequestTitle,
		pullrequestBody,
		pullrequestSourceBranch,
		pullrequestTargetBranch)

	logrus.Debugf("Repository:\t%s:%s", pullrequestRepositoryOwner, pullrequestRepositoryName)

	// Timeout api query after 30sec
	ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := g.client.PullRequests.Create(
		ctx,
		strings.Join([]string{
			pullrequestRepositoryOwner,
			pullrequestRepositoryName}, "/"),
		&opts,
	)

	if resp.Status > 400 {
		logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		return err
	}

	logrus.Infof("Gitea pullrequest successfully open on %q", pr.Link)

	return nil
}
