package gitea

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// CommitMessage represents conventional commit metadata as type or scope, used to generate the final commit message.
	CommitMessage commit.Commit `yaml:",omitempty"`
	// Directory specifies where the github repository is cloned on the local disk
	Directory string `yaml:",omitempty"`
	// Email specifies which emails to use when creating commits
	Email string `yaml:",omitempty"`
	// Force is used during the git push phase to run `git push --force`.
	Force bool `yaml:",omitempty"`
	// GPG key and passphrased used for commit signing
	GPG sign.GPGSpec `yaml:",omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// User specifies the user of the git commit messages
	User string `yaml:",omitempty"`
	// Branch specifies which Gitea repository branch to work on
	Branch string `yaml:",omitempty"`
}

// Gittea contains information to interact with Gitea api
type Gitea struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client     client.Client
	HeadBranch string
}

// New returns a new valid Gitea object.
func New(spec interface{}, pipelineID string) (*Gitea, error) {
	s := Spec{}
	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitea{}, nil
	}

	err = s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	c, err := client.New(client.Spec{
		URL:   s.URL,
		Token: s.Token,
	})

	if err != nil {
		return &Gitea{}, err
	}

	g := Gitea{
		Spec:       s,
		client:     c,
		HeadBranch: git.SanitizeBranchName(fmt.Sprintf("updatecli_%v", pipelineID)),
	}

	return &g, nil

}

// Retrieve git tags from a remote gitea repository
func (g *Gitea) SearchTags() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	references, resp, err := g.client.Git.ListTags(
		ctx,
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
		scm.ListOptions{
			URL:  g.Spec.URL,
			Page: 1,
			Size: 30,
		},
	)

	if err != nil {
		return nil, err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
	}

	for _, ref := range references {
		tags = append(tags, ref.Name)
	}

	return tags, nil
}

func (s *Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	err := s.ValidateClient()

	if err != nil {
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
