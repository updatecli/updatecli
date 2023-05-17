package gitlab

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with GitLab release
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
	// Branch specifies which GitLab repository branch to work on
	Branch string `yaml:",omitempty"`
}

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client           client.Client
	HeadBranch       string
	nativeGitHandler gitgeneric.GitHandler
}

// New returns a new valid GitLab object.
func New(spec interface{}, pipelineID string) (*Gitlab, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Gitlab{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitlab{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Gitlab{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "gitlab", s.Owner, s.Repository)
	}

	if len(s.Branch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Gitlab{}, err
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Gitlab{
		Spec:             s,
		client:           c,
		HeadBranch:       nativeGitHandler.SanitizeBranchName(fmt.Sprintf("updatecli_%v", pipelineID)),
		nativeGitHandler: nativeGitHandler,
	}

	g.setDirectory()

	return &g, nil

}

// Retrieve git tags from a remote gitlab repository
func (g *Gitlab) SearchTags() (tags []string, err error) {

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
		return fmt.Errorf("wrong gitlab configuration")
	}

	return nil
}
