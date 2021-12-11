package source

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Changelog string // Changelog hold the changelog description
	Result    string // Result store the source result after a source run. This variable can't be set by an updatecli configuration
	Output    string // Output contains the value retrieved from a source
	Config    Config // Config defines a source specifications
	Scm       *scm.ScmHandler
}

// Config struct defines a source configuration
type Config struct {
	resource.ResourceConfig `yaml:",inline"`
	// Deprecated in favor of Transformers on 2021/01/3
	Replaces Replacers
}

// Run execute actions defined by the source configuration
func (s *Source) Run() (err error) {
	// TODO-REFACTO: manage changelog
	// changelog := Changelog{}

	// TODO-REFACTO: call unmarshal in a constructor
	source, err := s.Config.ResourceConfig.Unmarshal()

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

		err = SCM.Init("", workingDir)

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

	// TODO-REFACTO: manage changelog
	// Retrieve changelog using default source output before
	// modifying its value with the transformer
	// if changelog != nil {
	// 	s.Changelog, err = changelog.Changelog(version.Version{
	// 		OriginalVersion: s.Output,
	// 		ParsedVersion:   s.Output,
	// 	})
	// 	if err != nil {
	// 		s.Result = result.FAILURE
	// 		// Changelog information are not important enough to fail a pipeline
	// 		logrus.Errorln(err)
	// 	}
	// }

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

	// Deprecated in favor of Transformers on 2021/01/3
	if len(s.Config.Replaces) > 0 {
		args := s.Config.Replaces.Unmarshal()

		r := strings.NewReplacer(args...)
		s.Output = (r.Replace(s.Output))
	}

	if len(s.Output) == 0 {
		s.Result = result.ATTENTION
	}

	return err
}
