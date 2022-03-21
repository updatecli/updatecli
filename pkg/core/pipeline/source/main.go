package source

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Changelog string // Changelog holds the changelog description
	Result    string // Result stores the source result after a source run. This variable can't be set by an updatecli configuration
	Output    string // Output contains the value retrieved from a source
	Config    Config // Config defines a source specifications
	Scm       *scm.ScmHandler
}

// Config struct defines a source configuration
type Config struct {
	resource.ResourceConfig `yaml:",inline"`
}

// Run execute actions defined by the source configuration
func (s *Source) Run() (err error) {
	source, err := resource.New(s.Config.ResourceConfig)
	if err != nil {
		s.Result = result.FAILURE
		return err
	}

	workingDir := ""

	if s.Scm != nil {

		SCM := *s.Scm

		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		err = SCM.Init(workingDir)

		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		err = SCM.Checkout()

		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		workingDir = SCM.GetDirectory()

	} else if s.Scm == nil {

		pwd, err := os.Getwd()
		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		workingDir = pwd
	}

	s.Output, err = source.Source(workingDir)
	s.Result = result.SUCCESS
	if err != nil {
		s.Result = result.FAILURE
		return err
	}

	// Once the source is executed, then it can retrieve its changelog
	// Any error means an empty changelog
	s.Changelog = source.Changelog()
	if s.Changelog == "" {
		logrus.Debugf("empty changelog found for the source %v", s)
	}

	if len(s.Config.Transformers) > 0 {
		s.Output, err = s.Config.Transformers.Apply(s.Output)
		if err != nil {
			s.Result = result.FAILURE
			return err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(s.Config.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release\n")
	}

	// Announce deprecation on 2021/01/31
	if len(s.Config.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release\n")
	}

	if len(s.Output) == 0 {
		s.Result = result.ATTENTION
	}

	return err
}
