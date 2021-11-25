package pullRequest

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
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
	Title   string      // Define the pullRequest Title
	Kind    string      // Define the pullRequest kind
	Spec    interface{} // Define specific parameters
	Targets []string    // DependsOnTargets defined a list of target related to the pullRequest
}

// PullRequest is a struct used by an updatecli pipeline.
type PullRequest struct {
	Title          string
	Changelog      string
	PipelineReport string
	Config         *Config
	PullRequest    PullRequestHandler
}

// Validate ensure that a pullRequest configuration has required parameters.
func (c *Config) Validate() error {
	errs := []error{}

	if len(c.Kind) == 0 {
		logrus.Errorln("Missing 'kind' values")
		errs = append(errs, errors.New("missing 'kind ' value"))
	}
	if c.Spec == nil {
		logrus.Errorln("Missing 'spec' value")
		errs = append(errs, errors.New("missing 'spec' value"))
	}

	if len(c.Targets) == 0 {
		logrus.Errorln("Missing at least one value in 'Targets'")
		errs = append(errs, errors.New("missing 'Targets' value"))
	}

	if len(errs) > 0 {
		return ErrWrongConfig
	}

	return nil

}

// New returns a new PullRequest object
func New(config *Config) (PullRequest, error) {

	p := PullRequest{
		Title:  config.Title,
		Config: config,
	}

	err := p.generatePullRequestHandler()
	if err != nil {
		return PullRequest{}, err
	}

	return p, nil

}

// Update a pullRequest object based on its configuration
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
		githubSpec := github.Spec{}

		err := mapstructure.Decode(p.Config.Spec, &githubSpec)
		if err != nil {
			return err
		}

		g, err := github.New(githubSpec)

		if err != nil {
			return err
		}

		p.PullRequest = &g

	default:
		logrus.Errorf("scm kind %q not supported", p.Config.Kind)
	}

	return nil
}
