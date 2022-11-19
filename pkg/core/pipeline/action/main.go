package pullrequest

import (
	"errors"
	"fmt"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	gitea "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/pullrequest"
	giteascm "github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
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
	// Title defines the pullRequest title
	Title string `yaml:",omitempty"`
	// Kind defines the pullRequest `kind` which affects accepted "spec" values
	Kind string `yaml:",omitempty"`
	// Spec defines parameters for a specific "kind"
	Spec interface{} `yaml:",omitempty"`
	// scmid references an scm configuration defined within the updatecli manifest
	ScmID string `yaml:",omitempty"`
	// !Deprecated in favor for `scmid`
	DeprecatedScmID string `yaml:"scmID,omitempty" jsonschema:"-"`
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

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

	if len(c.DeprecatedScmID) > 0 {

		switch len(c.ScmID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.ScmID = c.DeprecatedScmID
			c.DeprecatedScmID = ""
		default:
			logrus.Warningf("DeprecatedscmID: %q", c.DeprecatedScmID)
			logrus.Warningf("scmid: %q", c.ScmID)
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
				"scmID", "scmid", "scmID")
		}
	}

	if len(c.ScmID) == 0 {
		missingParameters = append(missingParameters, "scmid")
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
	case "gitea":
		pullRequestSpec := gitea.Spec{}

		if p.Scm.Config.Kind != "gitea" {
			return fmt.Errorf("scm of kind %q is not compatible with pullrequest of kind %q",
				p.Scm.Config.Kind,
				p.Config.Kind)
		}

		err := mapstructure.Decode(p.Config.Spec, &pullRequestSpec)
		if err != nil {
			return err
		}

		ge, ok := p.Scm.Handler.(*giteascm.Gitea)

		if !ok {
			return fmt.Errorf("scm is not of kind 'gitea'")
		}

		g, err := gitea.New(pullRequestSpec, ge)

		if err != nil {
			return err
		}

		p.Handler = &g

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

// JSONSchema implements the json schema interface to generate the "pullrequest" jsonschema
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := map[string]interface{}{
		"github": &github.PullRequestSpec{},
		"gitea":  &gitea.Spec{},
	}

	return jsonschema.GenerateJsonSchema(configAlias{}, anyOfSpec)
}
