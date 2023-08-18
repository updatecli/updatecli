package gitea

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
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	/*
		"commitMessage" is used to generate the final commit message.

		compatible:
			* scm

		remark:
			it's worth mentioning that the commit message settings will be applied to all targets linked to the same scm.
	*/
	CommitMessage commit.Commit `yaml:",omitempty"`
	/*
		"directory" defines the local path where the git repository is cloned.

		compatible:
			* scm

		remark:
			Unless you know what you are doing, it is recommended to use the default value.
			The reason is that Updatecli may automatically clean up the directory after a pipeline execution.

		default:
			/tmp/updatecli/github/<owner>/<repository>
	*/
	Directory string `yaml:",omitempty"`
	/*
		"email" defines the email used to commit changes.

		compatible:
			* scm

		default:
			default set to your global git configuration
	*/
	Email string `yaml:",omitempty"`
	/*
		"force" is used during the git push phase to run `git push --force`.

		compatible:
			* scm
	*/
	Force bool `yaml:",omitempty"`
	/*
		"gpg" specifies the GPG key and passphrased used for commit signing

		compatible:
			* scm
	*/
	GPG sign.GPGSpec `yaml:",omitempty"`
	/*
		"owner" defines the owner of a repository.

		compatible:
			* scm
	*/
	Owner string `yaml:",omitempty" jsonschema:"required"`
	/*
		repository specifies the name of a repository for a specific owner.

		compatible:
			* action
			* scm
	*/
	Repository string `yaml:",omitempty" jsonschema:"required"`
	/*
		"user" specifies the user associated with new git commit messages created by Updatecli

		compatible:
			* scm
	*/
	User string `yaml:",omitempty"`
	/*
		"branch" defines the git branch to work on.

		compatible:
			* scm

		default:
			main

		remark:
			depending on which resource references the Gitea scm, the behavior will be different.

			If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
			file(s) from that branch.

			If the scm is linked to target then Updatecli creates a new "working branch" based on the branch value.
			The working branch created by Updatecli looks like "updatecli_<pipelineID>".
			It is worth mentioning that it is not possible to by pass the working branch in the current situation.
			For more information, please refer to the following issue:
			https://github.com/updatecli/updatecli/issues/1139

			If you need to push changes to a specific branch, you must use the plugin "git" instead of this
	*/
	Branch string `yaml:",omitempty"`
}

// Gitea contains information to interact with Gitea api
type Gitea struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client           client.Client
	nativeGitHandler gitgeneric.GitHandler
	pipelineID       string
}

// New returns a new valid Gitea object.
func New(spec interface{}, pipelineID string) (*Gitea, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Gitea{}, err
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Gitea{}, err
	}

	err = clientSpec.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitea{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "gitea", s.Owner, s.Repository)
	}

	if len(s.Branch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Gitea{}, err
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Gitea{
		Spec:             s,
		client:           c,
		pipelineID:       pipelineID,
		nativeGitHandler: nativeGitHandler,
	}

	g.setDirectory()

	return &g, nil

}

// SearchTags retrieve git tags from a remote gitea repository
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
