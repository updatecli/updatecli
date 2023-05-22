package source

import (
	"errors"
	"os"
	"strings"

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
	Result result.Source
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

var (
	// ErrWrongConfig is returned when a condition spec has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong source configuration")
)

// Run execute actions defined by the source configuration
func (s *Source) Run() (err error) {

	source, err := resource.New(s.Config.ResourceConfig)
	if err != nil {
		s.Result.Result = result.FAILURE
		return err
	}

	workingDir := ""

	switch s.Scm == nil {
	case true:
		pwd, err := os.Getwd()
		if err != nil {
			s.Result.Result = result.FAILURE
			return err
		}

		workingDir = pwd
	case false:
		SCM := *s.Scm

		err = SCM.Checkout()
		if err != nil {
			s.Result.Result = result.FAILURE
			return err
		}

		workingDir = SCM.GetDirectory()
	}

	err = source.Source(workingDir, &s.Result)

	s.Result.Name = s.Config.ResourceConfig.Name
	s.Output = s.Result.Information

	if err != nil {
		s.Result.Result = result.FAILURE
		logrus.Errorf("%s %s", s.Result.Result, err)
		return err
	}

	logrus.Infof("%s %s", s.Result.Result, s.Result.Description)

	// Once the source is executed, then it can retrieve its changelog
	// Any error means an empty changelog
	s.Changelog = source.Changelog()
	if s.Changelog == "" {
		logrus.Debugln("empty changelog found for the source")
	}

	if len(s.Config.ResourceConfig.Transformers) > 0 {
		s.Output, err = s.Config.ResourceConfig.Transformers.Apply(s.Output)
		if err != nil {
			logrus.Errorf("%s %s", s.Result.Result, err)
			s.Result.Result = result.FAILURE
			return err
		}
	}

	if len(s.Output) == 0 && s.Result.Result == result.SUCCESS {
		logrus.Debugln("empty source detected")
	}

	return err
}

// JSONSchema implements the json schema interface to generate the "source" jsonschema.
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.AppendOneOfToJsonSchema(configAlias{}, anyOfSpec)
}

func (c *Config) Validate() error {
	gotError := false

	missingParameters := []string{}

	// Handle scmID deprecation
	if len(c.ResourceConfig.DeprecatedSCMID) > 0 {
		switch len(c.ResourceConfig.SCMID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.ResourceConfig.SCMID = c.ResourceConfig.DeprecatedSCMID
			c.ResourceConfig.DeprecatedSCMID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
				"scmID", "scmid", "scmID")
		}
	}

	// Validate that kind is set
	if len(c.ResourceConfig.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	// Handle depends_on deprecation
	if len(c.ResourceConfig.DeprecatedDependsOn) > 0 {
		switch len(c.ResourceConfig.DependsOn) == 0 {
		case true:
			logrus.Warningln("\"depends_on\" is deprecated in favor of \"dependson\".")
			c.ResourceConfig.DependsOn = c.ResourceConfig.DeprecatedDependsOn
			c.ResourceConfig.DeprecatedDependsOn = []string{}
		case false:
			logrus.Warningln("\"depends_on\" is ignored in favor of \"dependson\".")
			c.ResourceConfig.DeprecatedDependsOn = []string{}
		}
	}

	// Ensure kind is lowercase
	if c.ResourceConfig.Kind != strings.ToLower(c.ResourceConfig.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.ResourceConfig.Kind)
		c.ResourceConfig.Kind = strings.ToLower(c.ResourceConfig.Kind)
	}

	err := c.ResourceConfig.Transformers.Validate()
	if err != nil {
		logrus.Errorln(err)
		gotError = true
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
