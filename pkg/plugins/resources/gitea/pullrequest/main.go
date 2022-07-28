package pullrequest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
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
	client     client.Client
	HeadBranch string
}

// New returns a new valid Gitea object.
func New(spec interface{}) (*Gitea, error) {

	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Gitea{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitea{}, nil
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Gitea{}, err
	}

	err = clientSpec.Validate()

	if err != nil {
		return &Gitea{}, nil
	}

	s.Spec = clientSpec

	if len(s.TargetBranch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.TargetBranch = "main"
	}

	err = s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Gitea{}, err
	}

	return &Gitea{
		Spec:   s,
		client: c,
	}, nil

}

func (g *Gitea) CreatePullRequest(title, changelog, pipelineReport string) error {

	if len(g.Spec.Body) == 0 {
		g.Spec.Body = changelog + "\n" + pipelineReport
	}

	if len(g.Spec.Title) == 0 {
		g.Spec.Title = title
	}

	if len(g.Spec.SourceBranch) == 0 {
		g.Spec.SourceBranch = g.HeadBranch
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
		Title:  g.Spec.Title,
		Body:   g.Spec.Body,
		Source: g.Spec.SourceBranch,
		Target: g.Spec.TargetBranch,
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

func (s Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	err := s.Spec.Validate()

	if err != nil {
		logrus.Errorln(err)
		gotError = true
	}

	if len(s.Owner) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "owner")
	}

	if len(s.Repository) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "repository")
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong gitea configuration")
	}

	return nil
}
