package scm

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	"github.com/updatecli/updatecli/pkg/plugins/scms/stash"
)

type Scm struct {
	Config     *Config
	Handler    ScmHandler
	PipelineID string
}

var (
	// ErrWrongConfig is returned when a scm has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong scm configuration")
)

// ScmHandler is an interface offering common functions for a source control manager like git or github
type ScmHandler interface {
	Add(files []string) error
	Clone() (string, error)
	Checkout() error
	GetDirectory() (directory string)
	Commit(message string) error
	Clean() error
	Push() (bool, error)
	PushTag(tag string) error
	PushBranch(branch string) error
	GetChangedFiles(workingDir string) ([]string, error)
	IsRemoteBranchUpToDate() (bool, error)
	GetBranches() (sourceBranch, workingBranch, targetBranch string)
	GetURL() string
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
	case "stash":
		g, err := stash.New(s.Config.Spec, s.PipelineID)

		if err != nil {
			return err
		}

		s.Handler = g

	case "gitea":
		g, err := gitea.New(s.Config.Spec, s.PipelineID)

		if err != nil {
			return err
		}

		s.Handler = g

	case "gitlab":
		g, err := gitlab.New(s.Config.Spec, s.PipelineID)

		if err != nil {
			return err
		}

		s.Handler = g

	case "github":
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

	case "git":
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
	default:
		return fmt.Errorf("scm of kind %q is not supported", s.Config.Kind)
	}

	return nil
}
