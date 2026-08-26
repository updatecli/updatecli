package scm

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/scms/azuredevops"
	"github.com/updatecli/updatecli/pkg/plugins/scms/azuredevopssearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/bitbucket"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/githubsearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlabsearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/stash"
)

const (
	gitIdentifier         = "git"
	githubIdentifier      = "github"
	gitlabIdentifier      = "gitlab"
	azureDevopsIdentifier = "azuredevops"
	giteaIdentifier       = "gitea"
	stashIdentifier       = "stash"
	bitbucketIdentifier   = "bitbucket"

	githubsearchIdentifier      = "githubsearch"
	azuredevopssearchIdentifier = "azuredevopssearch"
	gitlabsearchIdentifier      = "gitlabsearch"
)

type Scm struct {
	Config     *Config
	Handler    ScmHandler
	PipelineID string
}

// ErrWrongConfig is returned when a scm has missing attributes which are mandatory
var ErrWrongConfig = errors.New("wrong scm configuration")

// ScmHandler is an interface offering common functions for a source control manager like git or github
type ScmHandler interface {
	Add(files []string) error
	CleanWorkingBranch() (bool, error)
	Clone() (string, error)
	Checkout() error
	GetDirectory() (directory string)
	Commit(ctx context.Context, message string) error
	Clean() error
	Push() (bool, error)
	PushTag(tag string) error
	PushBranch(branch string) error
	GetChangedFiles(workingDir string) ([]string, error)
	IsRemoteBranchUpToDate() (bool, error)
	IsRemoteWorkingBranchExist() (bool, error)
	GetBranches() (sourceBranch, workingBranch, targetBranch string)
	GetURL() string
	Summary() string
}

func New(config *Config, pipelineID string) (Scm, error) {
	s := Scm{
		Config:     config,
		PipelineID: pipelineID,
	}

	err := s.GenerateSCM()
	if err != nil {
		return Scm{}, err
	}

	return s, nil
}

// GenerateSCM populates the receiver's attribute "s.Handler" with the SCM implementation
// based on the "s.Conf" content
func (s *Scm) GenerateSCM() error {
	if s.Config.Disabled {
		return nil
	}

	switch s.Config.Kind {
	case azureDevopsIdentifier:
		g, err := azuredevops.New(s.Config.Spec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g

	case bitbucketIdentifier:
		g, err := bitbucket.New(s.Config.Spec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g

	case stashIdentifier:
		g, err := stash.New(s.Config.Spec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g

	case giteaIdentifier:
		g, err := gitea.New(s.Config.Spec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g

	case gitlabIdentifier:
		g, err := gitlab.New(s.Config.Spec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g

	case githubIdentifier:
		githubSpec := github.Spec{}

		err := mapstructure.Decode(s.Config.Spec, &githubSpec)
		if err != nil {
			return err
		}

		g, err := github.New(githubSpec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g

	case gitIdentifier:
		gitSpec := git.Spec{}

		err := mapstructure.Decode(s.Config.Spec, &gitSpec)
		if err != nil {
			return err
		}

		g, err := git.New(gitSpec, s.PipelineID)
		if err != nil {
			return err
		}

		s.Handler = g
	case githubsearch.Kind:
		// githubsearch scm kind is handled during engine preparation step
	case azuredevopssearch.Kind:
		// azuredevopssearch scm kind is handled during engine preparation step
	case gitlabsearch.Kind:
		// gitlabsearch scm kind is handled during engine preparation step
	default:
		return fmt.Errorf("scm of kind %q is not supported", s.Config.Kind)
	}

	return nil
}
