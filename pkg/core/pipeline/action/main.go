package action

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/reports"
	bitbucket "github.com/updatecli/updatecli/pkg/plugins/resources/bitbucket/pullrequest"
	gitea "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/pullrequest"
	gitlab "github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/mergerequest"
	stash "github.com/updatecli/updatecli/pkg/plugins/resources/stash/pullrequest"
	bitbucketscm "github.com/updatecli/updatecli/pkg/plugins/scms/bitbucket"
	giteascm "github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	gitlabscm "github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	stashscm "github.com/updatecli/updatecli/pkg/plugins/scms/stash"
)

const (
	gitlabIdentifier    = "gitlab"
	githubIdentifier    = "github"
	giteaIdentifier     = "gitea"
	stashIdentifier     = "stash"
	bitbucketIdentifier = "bitbucket"
)

// ErrWrongConfig is returned when an action has missing mandatory attributes.
var ErrWrongConfig = errors.New("wrong action configuration")

// ActionHandler interface defines required functions to be an action
type ActionHandler interface {
	CreateAction(report *reports.Action, resetDescription bool) error
	CleanAction(report *reports.Action) error
	CheckActionExist(report *reports.Action) error
}

// Config define action provided via an updatecli configuration
type Config struct {
	// Title defines the action title
	Title string `yaml:",omitempty"`
	// Kind defines the action `kind` which affects accepted "spec" values
	Kind string `yaml:",omitempty"`
	// Spec defines parameters for a specific "kind"
	Spec interface{} `yaml:",omitempty"`
	// scmid references a scm configuration defined within the updatecli manifest
	ScmID string `yaml:",omitempty"`
	// !Deprecated in favor of `scmid`
	DeprecatedScmID string `yaml:"scmID,omitempty" jsonschema:"-"`
	/*
		pipelineurl defines if a link to the Updatecli pipeline CI job should be added to the report.

		remarks:
		* Only available for GitHub Action, GitLab, Jenkins
	*/
	DisablePipelineURL bool `yaml:",omitempty"`
}

// Action is a struct used by an updatecli pipeline.
type Action struct {
	Title   string
	Config  Config
	Scm     *scm.Scm
	Handler ActionHandler
	Report  reports.Action
}

// Validate ensures that an action configuration has required parameters.
func (c *Config) Validate() (err error) {
	missingParameters := []string{}

	if c.Kind == "" {
		missingParameters = append(missingParameters, "kind")
	}

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

	/** Deprecated items **/
	if c.Kind == githubIdentifier {
		logrus.Warnf("The kind %q for actions is deprecated in favor of '%s/pullrequest'", githubIdentifier, githubIdentifier)
		c.Kind = "github/pullrequest"
	}
	if c.Kind == giteaIdentifier {
		logrus.Warnf("The kind %q for actions is deprecated in favor of '%s/pullrequest'", giteaIdentifier, giteaIdentifier)
		c.Kind = "gitea/pullrequest"
	}
	if c.DeprecatedScmID != "" {
		if c.ScmID == "" {
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.ScmID = c.DeprecatedScmID
		} else {
			logrus.Warningf("deprecatedscmID: %q", c.DeprecatedScmID)
			logrus.Warningf("scmid: %q", c.ScmID)
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
				"scmid", "deprecatedscmID", "deprecatedscmID")
		}
		c.DeprecatedScmID = ""
	}

	if c.ScmID == "" {
		missingParameters = append(missingParameters, "scmid")
	}

	if len(missingParameters) > 0 {
		err = fmt.Errorf("missing value for parameter(s) [%q]", strings.Join(missingParameters, ","))
	}

	return err
}

// New returns a new Action based on an action config and an scm
func New(config *Config, sourceControlManager *scm.Scm) (Action, error) {
	newAction := Action{
		Title:  config.Title,
		Config: *config,
		Scm:    sourceControlManager,
	}

	err := newAction.generateActionHandler()
	if err != nil {
		return Action{}, err
	}

	return newAction, nil
}

// Update updates an action object based on its configuration
func (a *Action) Update() error {

	err := a.generateActionHandler()
	if err != nil {
		return err
	}

	return nil
}

func (a *Action) generateActionHandler() error {
	// Don't forget to update the JSONSchema() method when adding/updating/removing a case
	switch a.Config.Kind {
	case "bitbucket/pullrequest", bitbucketIdentifier:
		actionSpec := bitbucket.Spec{}

		if a.Scm.Config.Kind != bitbucketIdentifier {
			return fmt.Errorf("scm of kind %q is not compatible with action of kind %q",
				a.Scm.Config.Kind,
				a.Config.Kind)
		}

		err := mapstructure.Decode(a.Config.Spec, &actionSpec)
		if err != nil {
			return err
		}

		ge, ok := a.Scm.Handler.(*bitbucketscm.Bitbucket)

		if !ok {
			return fmt.Errorf("scm is not of kind 'bitbucket'")
		}

		g, err := bitbucket.New(actionSpec, ge)
		if err != nil {
			return err
		}

		a.Handler = &g

	case "gitea/pullrequest", giteaIdentifier:
		actionSpec := gitea.Spec{}

		if a.Scm.Config.Kind != giteaIdentifier {
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

	case "gitlab/mergerequest", gitlabIdentifier:
		actionSpec := gitlab.Spec{}

		if a.Scm.Config.Kind != gitlabIdentifier {
			return fmt.Errorf("scm of kind %q is not compatible with action of kind %q",
				a.Scm.Config.Kind,
				a.Config.Kind)
		}

		err := mapstructure.Decode(a.Config.Spec, &actionSpec)
		if err != nil {
			return err
		}

		ge, ok := a.Scm.Handler.(*gitlabscm.Gitlab)

		if !ok {
			return fmt.Errorf("scm is not of kind 'gitlab'")
		}

		g, err := gitlab.New(actionSpec, ge)
		if err != nil {
			return err
		}

		a.Handler = &g

	case "github/pullrequest", githubIdentifier:
		actionSpec := github.ActionSpec{}

		if a.Scm.Config.Kind != githubIdentifier {
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

	case "stash/pullrequest", stashIdentifier:
		actionSpec := stash.Spec{}

		if a.Scm.Config.Kind != stashIdentifier {
			return fmt.Errorf("scm of kind %q is not compatible with action of kind %q",
				a.Scm.Config.Kind,
				a.Config.Kind)
		}

		err := mapstructure.Decode(a.Config.Spec, &actionSpec)
		if err != nil {
			return err
		}

		ge, ok := a.Scm.Handler.(*stashscm.Stash)

		if !ok {
			return fmt.Errorf("scm is not of kind 'stash'")
		}

		g, err := stash.New(actionSpec, ge)
		if err != nil {
			return err
		}

		a.Handler = &g

	default:
		logrus.Errorf("action of kind %q is not supported", a.Config.Kind)
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "action" jsonschema
func (Config) JSONSchema() *jschema.Schema {
	type configAlias Config

	anyOfSpec := map[string]interface{}{
		"github/pullrequest":    &github.ActionSpec{},
		"gitea/pullrequest":     &gitea.Spec{},
		"stash/pullrequest":     &stash.Spec{},
		"gitlab/mergerequest":   &gitlab.Spec{},
		"bitbucket/pullrequest": &bitbucket.Spec{},
	}

	return jsonschema.AppendOneOfToJsonSchema(configAlias{}, anyOfSpec)
}
