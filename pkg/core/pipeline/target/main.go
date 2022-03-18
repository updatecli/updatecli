package target

import (
	"errors"
	"fmt"
	"os"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrWrongConfig is returned when a target spec has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong target configuration")
)

// Target defines which file needs to be updated based on source output
type Target struct {
	// Holds the condition result after a target run.
	Result string
	// Holds the updatecli configuration of this target (usually the parsed YAML manifest excerpt that might be templatized)
	Config Config
	// Holds the dry run status (should the target apply the changes or only simulate)
	DryRun     bool
	WorkingDir string
	// Holds the change message (usually commit message associated to the target once run)
	Message string
	// Holds the list for changed files (usually to indicate SCM which files to stage/commit)
	Files []string
	// Holds the associated action
	Action *action.Action
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
	// disablesourceinput disables the mechanism to retrieve a default value from a source.
	DisableSourceInput bool `yaml:",omitempty"`
	// sourceid specifies where retrieving the default value
	SourceID string `yaml:",omitempty"`
	// disablesourceinput
}

// Check verifies if mandatory Targets parameters are provided and return false if not.
func (t *Target) Check() (bool, error) {
	ok := true
	required := []string{}

	if t.Config.Name == "" {
		required = append(required, "Name")
	}

	if len(required) > 0 {
		err := fmt.Errorf("%s Target parameter(s) required: [%v]", result.FAILURE, strings.Join(required, ","))
		return false, err
	}

	return ok, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string) error {
	var changed bool

	// TODO: prepare workingDir (1-copy)
	workingDir, err := t.initWorkingDir()
	if err != nil {
		return err
	}

	if len(t.Config.Transformers) > 0 {
		source, err = t.Config.Transformers.Apply(source)
		if err != nil {
			t.Result = result.FAILURE
			return err
		}
	}

	if t.DryRun {
		logrus.Infof("\n**Dry Run enabled**\n\n")
	}

	// Initialize the target resource based on its configuration
	target, err := resource.New(t.Config.ResourceConfig)
	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	changed, changedFiles, changeMessage, err := target.Target(source, workingDir, t.DryRun)
	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	t.Message = changeMessage
	t.Files = changedFiles

	if changed {
		t.Result = result.ATTENTION
	} else {
		t.Result = result.SUCCESS
	}

	// TODO: retrieve target diff/changed files

	// Execute the action's post-target step if needed
	// E.g. when not in dry run AND there is an action associated with this target (not empty)
	if !t.DryRun && t.Action != (&action.Action{}) {
		t.Result, err = t.Action.Handler.RunTarget(t.Message, t.Files)
		if err != nil {
			return err
		}
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "target" jsonschema.
func (Config) JSONSchema() *jschema.Schema {
	type configAlias Config

	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.GenerateJsonSchema(configAlias{}, anyOfSpec)
}

func (c *Config) Validate() error {
	// Handle scmID deprecation

	gotError := false

	missingParameters := []string{}

	// Validate that kind is set
	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

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

	// Handle sourceID deprecation
	if len(c.DeprecatedSourceID) > 0 {
		switch len(c.SourceID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "sourceID", "sourceid")
			c.SourceID = c.DeprecatedSourceID
			c.DeprecatedSourceID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusif, ignoring %q",
				"sourceID", "sourceid", "sourceID")
		}
	}

	err := c.Transformers.Validate()
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

func (t Target) initWorkingDir() (string, error) {
	// Delegate to the associated action if this target has one
	// Otherwise it's the current working directory (dry run or no action)
	if t.Action == (&action.Action{}) {
		return os.Getwd()
	}

	return t.Action.InitWorkingDir()
}
