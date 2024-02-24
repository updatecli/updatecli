package stash

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
	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with Bitbucket release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	/*
		"commitMessage" is used to generate the final commit message.

		compatible:
			* scm

		remark:
			it's worth mentioning that the commit message settings is applied to all targets linked to the same scm.
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
			depending on which resource references the Stash scm, the behavior will be different.

			If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
			file(s) from that branch.

			If the scm is linked to target then Updatecli creates a new "working branch" based on the branch value.
			The working branch created by Updatecli looks like "updatecli_<pipelineID>".
			It is worth mentioning that it is not possible to bypass the working branch in the current situation.
			For more information, please refer to the following issue:
			https://github.com/updatecli/updatecli/issues/1139

			If you need to push changes to a specific branch, you must use the plugin "git" instead of this
	*/
	Branch string `yaml:",omitempty"`
	// Whether to checkout submodules: `true` to checkout submodules or `false` to skip.
	Submodules *bool `yaml:",omitempty"`
	// workingBranch defines if Updatecli should use a temporary branch to work on.
	// If set to `true`, Updatecli create a temporary branch to work on, based on the branch value.
	//
	// default: true
	WorkingBranch *bool `yaml:",omitempty"`
}

// Stash contains information to interact with Stash api
type Stash struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client           client.Client
	pipelineID       string
	nativeGitHandler gitgeneric.GitHandler
	workingBranch    bool
}

// New returns a new valid Bitbucket object.
func New(spec interface{}, pipelineID string) (*Stash, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Stash{}, err
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Stash{}, err
	}

	err = clientSpec.Validate()

	if err != nil {
		return &Stash{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Stash{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Stash{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "stash", s.Owner, s.Repository)
	}

	if len(s.Branch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
	}

	workingBranch := true
	if s.WorkingBranch != nil {
		workingBranch = *s.WorkingBranch
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Stash{}, err
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Stash{
		Spec:             s,
		client:           c,
		pipelineID:       pipelineID,
		nativeGitHandler: nativeGitHandler,
		workingBranch:    workingBranch,
	}

	g.setDirectory()

	return &g, nil

}

// Retrieve git tags from a remote bitbucket repository
func (s *Stash) SearchTags() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	references, resp, err := s.client.Git.ListTags(
		ctx,
		strings.Join([]string{s.Spec.Owner, s.Spec.Repository}, "/"),
		scm.ListOptions{
			URL:  s.Spec.URL,
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

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong bitbucket configuration")
	}

	return nil
}
