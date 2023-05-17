package target

import (
	"errors"
	"fmt"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrWrongConfig is returned when a target spec has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong target configuration")
)

// Target defines which file needs to be updated based on source output
type Target struct {
	// Result store the condition result after a target run.
	Result result.Target
	Config Config
	Commit bool
	Push   bool
	Clean  bool
	DryRun bool
	Scm    *scm.ScmHandler
}

// Config defines target parameters
type Config struct {
	resource.ResourceConfig `yaml:",inline"`
	// ReportTitle contains the updatecli reports title for sources and conditions run
	ReportTitle string `yaml:",omitempty"`
	// ReportBody contains the updatecli reports body for sources and conditions run
	ReportBody string `yaml:",omitempty"`
	// ! Deprecated - please use all lowercase `sourceid`
	DeprecatedSourceID string `yaml:"sourceID,omitempty" jsonschema:"-"`
	// disablesourceinput disables the mechanism to retrieve a default value from a source. For example, if true, source information like changelog will not be accessible for a github/pullrequest action.
	DisableSourceInput bool `yaml:",omitempty"`
	// sourceid specifies where retrieving the default value
	SourceID string `yaml:",omitempty"`
}

// Check verifies if mandatory Targets parameters are provided and return false if not.
func (t *Target) Check() (bool, error) {
	ok := true
	required := []string{}

	if t.Config.ResourceConfig.Name == "" {
		required = append(required, "Name")
	}

	if len(required) > 0 {
		err := fmt.Errorf("%s Target parameter(s) required: [%v]", result.FAILURE, strings.Join(required, ","))
		return false, err
	}

	return ok, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string, o *Options) (err error) {

	var changed bool

	if len(t.Config.ResourceConfig.Transformers) > 0 {
		source, err = t.Config.ResourceConfig.Transformers.Apply(source)
		if err != nil {
			t.Result.Result = result.FAILURE
			return err
		}
	}

	if o.DryRun {
		logrus.Infof("\n**Dry Run enabled**\n\n")
	}

	target, err := resource.New(t.Config.ResourceConfig)
	if err != nil {
		t.Result.Result = result.FAILURE
		return err
	}

	// Ensure the result named contains the up to date target name
	// after templating
	t.Result.Name = t.Config.ResourceConfig.Name

	// If no scm configuration provided then stop early
	if t.Scm == nil {

		err = target.Target(source, nil, o.DryRun, &t.Result)
		if err != nil {
			t.Result.Description = "something went wrong during pipeline execution"
			t.Result.Result = result.FAILURE
			return err
		}

		// Could be improve to show attention description in yellow, success in green, failure in red
		logrus.Infof("%s - %s", t.Result.Result, t.Result.Description)

		return nil

	}

	var message string
	var files []string

	_, err = t.Check()
	if err != nil {
		t.Result.Result = result.FAILURE
		return err
	}

	s := *t.Scm

	if err = s.Checkout(); err != nil {
		t.Result.Result = result.FAILURE
		return err
	}

	err = target.Target(source, s, o.DryRun, &t.Result)
	if err != nil {
		t.Result.Result = result.FAILURE
		return err
	}

	// Could be improve to show attention description in yellow, success in green, failure in red
	logrus.Infof("%s - %s", t.Result.Result, t.Result.Description)

	isRemoteBranchUpToDate, err := s.IsRemoteBranchUpToDate()
	if err != nil {
		t.Result.Result = result.FAILURE
		return err
	}

	if !t.Result.Changed {
		if isRemoteBranchUpToDate {
			return nil
		}

		logrus.Infof("\n\u26A0 While nothing change in the current pipeline run, according to the git history, some commits will be pushed\n")
	}

	if !o.DryRun {
		if changed {
			if t.Result.Description == "" {
				t.Result.Result = result.FAILURE
				return fmt.Errorf("target has no change message")
			}

			if len(t.Result.Files) == 0 {
				t.Result.Result = result.FAILURE
				return fmt.Errorf("no changed file to commit")
			}

			if o.Commit {
				if err := s.Add(files); err != nil {
					t.Result.Result = result.FAILURE
					return err
				}

				if err = s.Commit(message); err != nil {
					t.Result.Result = result.FAILURE
					return err
				}
			}
		}

		if o.Push {
			if err := s.Push(); err != nil {
				t.Result.Result = result.FAILURE
				return err
			}
		}
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "target" jsonschema.
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.AppendOneOfToJsonSchema(configAlias{}, anyOfSpec)
}

func (c *Config) Validate() error {
	// Handle scmID deprecation

	gotError := false

	missingParameters := []string{}

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

	// Handle sourceID deprecation
	if len(c.DeprecatedSourceID) > 0 {
		switch len(c.SourceID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "sourceID", "sourceid")
			c.SourceID = c.DeprecatedSourceID
			c.DeprecatedSourceID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
				"sourceID", "sourceid", "sourceID")
		}
	}

	err := c.ResourceConfig.Transformers.Validate()
	if err != nil {
		logrus.Errorln(err)
		gotError = true
	}

	if len(c.SourceID) > 0 && c.DisableSourceInput {
		logrus.Errorln("disablesourceinput is incompatible with sourceid, ignoring the latter")
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
