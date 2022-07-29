package pullrequest

import (
	"context"
	"errors"
	"fmt"
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

	err = clientSpec.Sanitize()
	if err != nil {
		return Gitea{}, err
	}

	if len(clientSpec.Token) == 0 && len(ge.Spec.Token) > 0 {
		clientSpec.Token = ge.Spec.Token
	}

	if len(clientSpec.URL) == 0 && len(ge.Spec.URL) > 0 {
		clientSpec.URL = ge.Spec.URL
	}

	if len(clientSpec.URL) == 0 {
		return Gitea{}, errors.New("error gitea server url not found")
	}

	if len(clientSpec.Username) == 0 && len(ge.Spec.Username) > 0 {
		clientSpec.Username = ge.Spec.Username
	}

	s.Spec = clientSpec

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

	pullrequestBody := changelog + "\n" + pipelineReport
	pullrequestTitle := title
	pullrequestSourceBranch := g.ge.HeadBranch
	pullrequestTargetBranch := g.ge.Spec.Branch

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
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
		optsSearch,
	)

	if err != nil {
		return err
	}

	if resp.Status > 400 {
		logrus.Debugf("Gitea Api Response: %v", resp)
	}

	for _, p := range pullrequests {
		if p.Source == g.Spec.SourceBranch && p.Target == g.Spec.TargetBranch {
			logrus.Infof("%s A pullrequest from branch %q to branch %q is already opened, nothing else to do",
				result.SUCCESS,
				g.Spec.SourceBranch,
				g.Spec.TargetBranch)
			return nil
		}
	}

	opts := scm.PullRequestInput{
		Title:  pullrequestTitle,
		Body:   pullrequestBody,
		Source: pullrequestSourceBranch,
		Target: pullrequestTargetBranch,
	}

	// Timeout api query after 30sec
	ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := g.client.PullRequests.Create(
		ctx,
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
		&opts,
	)

	if err != nil {
		return err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
		return fmt.Errorf("error from api: %v", resp.Status)
	}

	logrus.Infof("Gitea pullrequest successfully open on %q", pr.Link)

	return nil
}
