package pullrequest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/github"
)

var (
	// ErrWrongConfig is returned when a pullrequest has missing mandatory attributes.
	ErrWrongConfig = errors.New("wrong pull request configuration")
)

// PullRequestHandler interface defines required functions to be an pullRequest
type PullRequestHandler interface {
	CreatePullRequest(title, changelog, pipelineReport string) error
}

// Config define pullRequest provided via an updatecli configuration
type Config struct {
	Title   string      // Defines the pullRequest Title
	Kind    string      // Defines the pullRequest kind
	Spec    interface{} // Defines specific parameters
	ScmID   string      `yaml:"scmID"` // References a scm configuration
	Targets []string    // DependsOnTargets defines a list of target related to the pullRequest
}

// PullRequest is a struct used by an updatecli pipeline.
type PullRequest struct {
	Title          string
	Changelog      string
	PipelineReport string
	Config         Config
	Scm            *scm.Scm
	Handler        PullRequestHandler
}

// Validate ensures that a pullRequest configuration has required parameters.
func (c *Config) Validate() (err error) {

	missingParameters := []string{}

	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	if len(c.Targets) == 0 {
		missingParameters = append(missingParameters, "targets")
	}

	if len(c.ScmID) == 0 {
		missingParameters = append(missingParameters, "scmID")
	}

	if len(missingParameters) > 0 {
		err = fmt.Errorf("missing value for parameter(s) [%q]", strings.Join(missingParameters, ","))
	}

	return err

}

// New returns a new PullRequest based on a pullrequest config and an scm
func New(config *Config, sourceControlManager *scm.Scm) (PullRequest, error) {

	p := PullRequest{
		Title:  config.Title,
		Config: *config,
		Scm:    sourceControlManager,
	}

	err := p.generatePullRequestHandler()
	if err != nil {
		return PullRequest{}, err
	}

	return p, nil

}

// Update updates a pullRequest object based on its configuration
func (p *PullRequest) Update() error {
	p.Title = p.Config.Title

	err := p.generatePullRequestHandler()
	if err != nil {
		return err
	}

	return nil
}

func (p *PullRequest) generatePullRequestHandler() error {

	switch p.Config.Kind {
	case "github":
		pullRequestSpec := github.PullRequestSpec{}

		if p.Scm.Config.Kind != "github" {
			return fmt.Errorf("scm of kind %q is not compatible with pullrequest of kind %q",
				p.Scm.Config.Kind,
				p.Config.Kind)
		}

		err := mapstructure.Decode(p.Config.Spec, &pullRequestSpec)
		if err != nil {
			return err
		}

		gh, ok := p.Scm.Handler.(*github.Github)

		if !ok {
			return fmt.Errorf("scm is not of kind 'github'")
		}

		g, err := github.NewPullRequest(pullRequestSpec, gh)

		if err != nil {
			return err
		}

		p.Handler = &g

	default:
		logrus.Errorf("scm of kind %q is not supported", p.Config.Kind)
	}

	return nil
}
