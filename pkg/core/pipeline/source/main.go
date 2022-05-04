package source

import (
	"os"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	// Changelog holds the changelog description
	Changelog string
	// Result stores the source result after a source run.
	Result string
	// Output contains the value retrieved from a source
	Output string
	// Config defines a source specifications
	Config Config
	Scm    *scm.ScmHandler
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

	if len(s.Output) == 0 {
		s.Result = result.ATTENTION
	}

	return err
}

// JSONSchema implements the json schema interface to generate the "source" jsonschema.
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.GenerateJsonSchema(configAlias{}, anyOfSpec)
}

func (c *Config) Validate() error {
	// Handle scmID deprecation
	if len(c.DeprecatedSCMID) > 0 {
		switch len(c.SCMID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.SCMID = c.DeprecatedSCMID
			c.DeprecatedSCMID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusif, ignoring %q",
				"scmID", "scmid", "scmID")
		}
	}

	err := c.Transformers.Validate()
	if err != nil {
		return err
	}

	return nil
}
