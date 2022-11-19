package action

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
	// ErrWrongConfig is returned when an action has missing mandatory attributes.
	ErrWrongConfig = errors.New("wrong action configuration")
)

// ActionHandler interface defines required functions to be an action
type ActionHandler interface {
	CreateAction(title, changelog, pipelineReport string) error
}

// Config define action provided via an updatecli configuration
type Config struct {
	// Title defines the action title
	Title string `yaml:",omitempty"`
	// Kind defines the action `kind` which affects accepted "spec" values
	Kind string `yaml:",omitempty"`
	// Spec defines parameters for a specific "kind"
	Spec interface{} `yaml:",omitempty"`
	// scmid references an scm configuration defined within the updatecli manifest
	ScmID string `yaml:",omitempty"`
	// !Deprecated in favor for `scmid`
	DeprecatedScmID string `yaml:"scmID,omitempty" jsonschema:"-"`
}

// Action is a struct used by an updatecli pipeline.
type Action struct {
	Title          string
	Changelog      string
	PipelineReport string
	Config         Config
	Scm            *scm.Scm
	Handler        ActionHandler
}

// Validate ensures that an action configuration has required parameters.
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

// New returns a new Action based on an action config and an scm
func New(config *Config, sourceControlManager *scm.Scm) (Action, error) {

	p := Action{
		Title:  config.Title,
		Config: *config,
		Scm:    sourceControlManager,
	}

	err := p.generateActionHandler()
	if err != nil {
		return Action{}, err
	}

	return p, nil

}

// Update updates an action object based on its configuration
func (p *Action) Update() error {
	p.Title = p.Config.Title

	err := p.generateActionHandler()
	if err != nil {
		return err
	}

	return nil
}

func (a *Action) generateActionHandler() error {

	switch a.Config.Kind {
	case "gitea":
		actionSpec := gitea.Spec{}

		if a.Scm.Config.Kind != "gitea" {
			return fmt.Errorf("scm of kind %q is not compatible with action of kind %q",
				a.Scm.Config.Kind,
				a.Config.Kind)
		}

		err := mapstructure.Decode(a.Config.Spec, &actionSpec)
		if err != nil {
			return err
		}

		ge, ok := a.Scm.Handler.(*giteascm.Gitea)

		if !ok {
			return fmt.Errorf("scm is not of kind 'gitea'")
		}

		g, err := gitea.New(actionSpec, ge)

		if err != nil {
			return err
		}

		a.Handler = &g

	case "github":
		actionSpec := github.ActionSpec{}

		if a.Scm.Config.Kind != "github" {
			return fmt.Errorf("scm of kind %q is not compatible with action of kind %q",
				a.Scm.Config.Kind,
				a.Config.Kind)
		}

		err := mapstructure.Decode(a.Config.Spec, &actionSpec)
		if err != nil {
			return err
		}

		gh, ok := a.Scm.Handler.(*github.Github)

		if !ok {
			return fmt.Errorf("scm is not of kind 'github'")
		}

		g, err := github.NewAction(actionSpec, gh)

		if err != nil {
			return err
		}

		a.Handler = &g

	default:
		logrus.Errorf("scm of kind %q is not supported", a.Config.Kind)
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "action" jsonschema
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := map[string]interface{}{
		"github": &github.ActionSpec{},
		"gitea":  &gitea.Spec{},
	}

	return jsonschema.GenerateJsonSchema(configAlias{}, anyOfSpec)
}
