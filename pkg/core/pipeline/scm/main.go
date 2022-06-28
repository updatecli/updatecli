package scm

import (
	"errors"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

type Config struct {
	// Kind specifies the scm kind
	Kind string `yaml:",omitempty"`
	// Spec specifies the scm specification
	Spec interface{} `jsonschema:"type=object" yaml:",omitempty"`
}

type Scm struct {
	Config     *Config
	Handler    ScmHandler
	PipelineID string
}

var (
	// ErrWrongConfig is returned when a scm has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong scm configuration")
)

// Scm is an interface that offers common functions for a source control manager like git or github
type ScmHandler interface {
	Add(files []string) error
	Clone() (string, error)
	Checkout() error
	GetDirectory() (directory string)
	Push() error
	Commit(message string) error
	Clean() error
	PushTag(tag string) error
	GetChangedFiles(workingDir string) ([]string, error)
}

func (c *Config) Validate() error {
	gotError := false
	missingParameters := []string{}

	// Validate that kind is set
	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	if c.Spec == nil {
		missingParameters = append(missingParameters, "spec")
	}

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing value for parameter(s) [%q]", strings.Join(missingParameters, ","))
		gotError = true
	}

	if gotError {
		return ErrWrongConfig
	}
	return nil
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

func (s *Scm) GenerateSCM() error {

	switch s.Config.Kind {
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

		g, err := git.New(gitSpec)

		if err != nil {
			return err
		}

		s.Handler = g
	default:
		logrus.Errorf("scm of kind %q is not supported", s.Config.Kind)
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "scm" jsonschema
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := map[string]interface{}{
		"git":    &git.Spec{},
		"github": &github.Spec{},
	}

	return jsonschema.GenerateJsonSchema(configAlias{}, anyOfSpec)
}
