package source

import (
	"bytes"
	"errors"
	"io"
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
	// Result stores the source result after a source run.
	Result result.Source
	// Output contains the value retrieved from a source
	Output string
	// OriginalOutput contains the raw value retrieved from a source before the transformation
	OriginalOutput string
	// Config defines a source specifications
	Config Config
	// Scm stores scm information
	Scm *scm.ScmHandler
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

	var consoleOutput bytes.Buffer
	// By default logrus logs to stderr, so I guess we want to keep this behavior...
	logrus.SetOutput(io.MultiWriter(os.Stdout, &consoleOutput))
	/*
		The last defer will be executed first,
		so in this case we want to first save the console output
		before setting back the logrus output to stdout.
	*/
	// By default logrus logs to stdout and we want to keep this behavior.
	defer logrus.SetOutput(os.Stdout)
	defer s.Result.SetConsoleOutput(&consoleOutput)

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

		s.Result.Scm.URL = SCM.GetURL()
		s.Result.Scm.Branch.Source, s.Result.Scm.Branch.Working, s.Result.Scm.Branch.Target = SCM.GetBranches()

		err = SCM.Checkout()
		if err != nil {
			s.Result.Result = result.FAILURE
			return err
		}

		workingDir = SCM.GetDirectory()
	}

	err = source.Source(workingDir, &s.Result)

	s.Output = s.Result.Information
	s.OriginalOutput = s.Result.Information

	if err != nil {
		s.Result.Result = result.FAILURE
		logrus.Errorf("%s %s", s.Result.Result, err)
		return err
	}

	logrus.Infof("%s %s", s.Result.Result, s.Result.Description)

	if len(s.Config.Transformers) > 0 {
		s.Output, err = s.Config.Transformers.Apply(s.Output)
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

// Validate checks if a source configuration is valid
func (c *Config) Validate() error {
	gotError := false

	missingParameters := []string{}

	// Handle scmID deprecation
	if len(c.DeprecatedSCMID) > 0 {
		switch len(c.SCMID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.SCMID = c.DeprecatedSCMID
			c.DeprecatedSCMID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
				"scmID", "scmid", "scmID")
		}
	}

	// Validate that kind is set
	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	// Handle depends_on deprecation
	if len(c.DeprecatedDependsOn) > 0 {
		switch len(c.DependsOn) == 0 {
		case true:
			logrus.Warningln("\"depends_on\" is deprecated in favor of \"dependson\".")
			c.DependsOn = c.DeprecatedDependsOn
			c.DeprecatedDependsOn = []string{}
		case false:
			logrus.Warningln("\"depends_on\" is ignored in favor of \"dependson\".")
			c.DeprecatedDependsOn = []string{}
		}
	}

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

	err := c.Transformers.Validate()
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
