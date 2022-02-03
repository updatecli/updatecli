package scm

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/git"
	"github.com/updatecli/updatecli/pkg/plugins/github"
)

type Config struct {
	Kind string
	Spec interface{}
}

type Scm struct {
	Config  *Config
	Handler ScmHandler
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
	Init(source string, pipelineID string) error
	Push() error
	Commit(message string) error
	Clean() error
	PushTag(tag string) error
	GetChangedFiles(workingDir string) ([]string, error)
}

func (c *Config) Validate() error {
	errs := []error{}

	if len(c.Kind) == 0 {
		logrus.Errorln("Missing 'kind' values")
		errs = append(errs, errors.New("missing 'kind' value"))
	}
	if c.Spec == nil {
		logrus.Errorln("Missing 'spec' value")
		errs = append(errs, errors.New("missing 'spec' value"))
	}

	if len(errs) > 0 {
		return ErrWrongConfig
	}
	return nil
}

func New(config *Config) (Scm, error) {

	s := Scm{
		Config: config,
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

		g, err := github.New(githubSpec)

		if err != nil {
			return err
		}

		s.Handler = g

	case "git":
		g := git.Git{}

		err := mapstructure.Decode(s.Config.Spec, &g)
		if err != nil {
			return err
		}

		s.Handler = &g
	default:
		logrus.Errorf("scm of kind %q is not supported", s.Config.Kind)
	}

	return nil
}
